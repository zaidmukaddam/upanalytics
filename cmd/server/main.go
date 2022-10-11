package main

import (
	"flag"
	"log"

	"upanalytics/internal/config"
	"upanalytics/internal/crawler"
	"upanalytics/internal/datastore"
	"upanalytics/internal/export"
	"upanalytics/internal/http"
	"upanalytics/internal/issue"
	"upanalytics/internal/project"
	"upanalytics/internal/projectview"
	"upanalytics/internal/pubsub"
	"upanalytics/internal/report"
	"upanalytics/internal/user"
)

func main() {
	var fname string
	var path string

	flag.StringVar(&fname, "c", "config", "Specify config filename. Default is config")
	flag.StringVar(&path, "p", ".", "Specify config path. Default is current directory")
	flag.Parse()

	config, err := config.NewConfig(path, fname)
	if err != nil {
		log.Fatalf("Error loading config: %v\n", err)
	}

	ds, err := datastore.NewDataStore(config.DB)
	if err != nil {
		log.Fatalf("Error creating new datastore: %v\n", err)
	}

	err = ds.Migrate()
	if err != nil {
		log.Fatalf("Error running migrations: %v\n", err)
	}

	broker := pubsub.NewBroker()

	services := &http.Services{
		UserService:        user.NewService(ds),
		ProjectService:     project.NewService(ds),
		CrawlerService:     crawler.NewService(ds, broker, config.Crawler),
		IssueService:       issue.NewService(ds),
		ReportService:      report.NewService(ds),
		ReportManager:      newReportManager(ds),
		ProjectViewService: projectview.NewService(ds),
		PubSubBroker:       broker,
		ExportService:      export.NewExporter(ds),
	}

	server := http.NewApp(
		config.HTTPServer,
		services,
	)

	server.Run()
}

// Create a new ReportManager with all available issue reporters.
func newReportManager(ds *datastore.Datastore) *issue.ReportManager {
	rm := issue.NewReportManager(ds)

	rm.AddReporter(ds.Find30xPageReports, issue.Error30x)
	rm.AddReporter(ds.Find40xPageReports, issue.Error40x)
	rm.AddReporter(ds.Find50xPageReports, issue.Error50x)
	rm.AddReporter(ds.FindPageReportsWithDuplicatedTitle, issue.ErrorDuplicatedTitle)
	rm.AddReporter(ds.FindPageReportsWithDuplicatedDescription, issue.ErrorDuplicatedDescription)
	rm.AddReporter(ds.FindPageReportsWithEmptyTitle, issue.ErrorEmptyTitle)
	rm.AddReporter(ds.FindPageReportsWithShortTitle, issue.ErrorShortTitle)
	rm.AddReporter(ds.FindPageReportsWithLongTitle, issue.ErrorLongTitle)
	rm.AddReporter(ds.FindPageReportsWithEmptyDescription, issue.ErrorEmptyDescription)
	rm.AddReporter(ds.FindPageReportsWithShortDescription, issue.ErrorShortDescription)
	rm.AddReporter(ds.FindPageReportsWithLongDescription, issue.ErrorLongDescription)
	rm.AddReporter(ds.FindPageReportsWithLittleContent, issue.ErrorLittleContent)
	rm.AddReporter(ds.FindImagesWithNoAlt, issue.ErrorImagesWithNoAlt)
	rm.AddReporter(ds.FindRedirectChains, issue.ErrorRedirectChain)
	rm.AddReporter(ds.FindPageReportsWithoutH1, issue.ErrorNoH1)
	rm.AddReporter(ds.FindPageReportsWithNoLangAttr, issue.ErrorNoLang)
	rm.AddReporter(ds.FindPageReportsWithHTTPLinks, issue.ErrorHTTPLinks)
	rm.AddReporter(ds.FindMissingHrelangReturnLinks, issue.ErrorHreflangsReturnLink)
	rm.AddReporter(ds.TooManyLinks, issue.ErrorTooManyLinks)
	rm.AddReporter(ds.InternalNoFollowLinks, issue.ErrorInternalNoFollow)
	rm.AddReporter(ds.FindExternalLinkWitoutNoFollow, issue.ErrorExternalWithoutNoFollow)
	rm.AddReporter(ds.FindCanonicalizedToNonCanonical, issue.ErrorCanonicalizedToNonCanonical)
	rm.AddReporter(ds.FindRedirectLoops, issue.ErrorRedirectLoop)
	rm.AddReporter(ds.FindNotValidHeadingsOrder, issue.ErrorNotValidHeadings)
	rm.AddReporter(ds.FindHreflangsToNonCanonical, issue.HreflangToNonCanonical)
	rm.AddReporter(ds.InternalNoFollowIndexableLinks, issue.ErrorInternalNoFollowIndexable)
	rm.AddReporter(ds.NoIndexable, issue.ErrorNoIndexable)
	rm.AddReporter(ds.HreflangNoindexable, issue.HreflangNoindexable)
	rm.AddReporter(ds.FindBlockedByRobotstxt, issue.ErrorBlocked)
	rm.AddReporter(ds.FindOrphanPages, issue.ErrorOrphan)

	return rm
}
