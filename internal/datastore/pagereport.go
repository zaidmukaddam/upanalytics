package datastore

import (
	"log"
	"math"

	"upanalytics/internal/crawler"
	"upanalytics/internal/issue"
)

func (ds *Datastore) SavePageReport(r *crawler.PageReport, cid int64) {
	urlHash := Hash(r.URL)
	var redirectHash string
	if r.RedirectURL != "" {
		redirectHash = Hash(r.RedirectURL)
	}

	query := `
		INSERT INTO pagereports (
			crawl_id,
			url,
			url_hash,
			scheme,
			redirect_url,
			redirect_hash,
			refresh,
			status_code,
			content_type,
			media_type,
			lang,
			title,
			description,
			robots,
			noindex,
			canonical,
			h1,
			h2,
			words,
			size,
			valid_headings,
			robotstxt_blocked,
			crawled
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := ds.db.Prepare(query)
	if err != nil {
		log.Printf("saveReport: %v\n", err)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		cid,
		r.URL,
		urlHash,
		r.ParsedURL.Scheme,
		r.RedirectURL,
		redirectHash,
		r.Refresh,
		r.StatusCode,
		r.ContentType,
		r.MediaType,
		r.Lang,
		r.Title,
		r.Description,
		r.Robots,
		r.Noindex,
		r.Canonical,
		r.H1,
		r.H2,
		r.Words,
		len(r.Body),
		r.ValidHeadings,
		r.BlockedByRobotstxt,
		r.Crawled,
	)

	if err != nil {
		log.Printf("Error in SavePageReport\nCID: %v\n Report: %+v\nError: %+v\n", cid, r, err)
		return
	}

	lid, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		return
	}

	if len(r.Links) > 0 {
		sqlString := "INSERT INTO links (pagereport_id, crawl_id, url, scheme, rel, nofollow, text, url_hash) values "
		v := []interface{}{}
		for _, l := range r.Links {
			hash := Hash(l.URL)
			sqlString += "(?, ?, ?, ?, ?, ?, ?, ?),"
			v = append(v, lid, cid, l.URL, l.ParsedURL.Scheme, l.Rel, l.NoFollow, l.Text, hash)
		}
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, err := ds.db.Prepare(sqlString)
		if err != nil {
			log.Printf("saveReport links: %v\n", err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(v...)
		if err != nil {
			log.Printf("Error in SavePageReport\nCID: %v\n Links: %+v\nError: %+v\n", cid, v, err)
		}
	}

	if len(r.ExternalLinks) > 0 {
		sqlString := "INSERT INTO external_links (pagereport_id, crawl_id, url, rel, nofollow, text, sponsored, ugc) values "
		v := []interface{}{}
		for _, l := range r.ExternalLinks {
			sqlString += "(?, ?, ?, ?, ?, ?, ?, ?),"
			v = append(v, lid, cid, l.URL, l.Rel, l.NoFollow, l.Text, l.Sponsored, l.UGC)
		}
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, err := ds.db.Prepare(sqlString)
		if err != nil {
			log.Println(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(v...)
		if err != nil {
			log.Printf("Error in SavePageReport\nCID: %v\n Links: %+v\nError: %+v\n", cid, v, err)
		}
	}

	if len(r.Hreflangs) > 0 {
		sqlString := "INSERT INTO hreflangs (pagereport_id, crawl_id, from_lang, to_url, to_lang, from_hash, to_hash) values "
		v := []interface{}{}
		for _, h := range r.Hreflangs {
			sqlString += "(?, ?, ?, ?, ?, ?, ?),"
			v = append(v, lid, cid, r.Lang, h.URL, h.Lang, Hash(r.URL), Hash(h.URL))
		}
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, _ := ds.db.Prepare(sqlString)
		defer stmt.Close()

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Printf("savePageReport\nCID: %v\n Hreflangs: %+v\nError: %+v\n", cid, v, err)
		}
	}

	if len(r.Images) > 0 {
		sqlString := "INSERT INTO images (pagereport_id, url, alt, crawl_id) values "
		v := []interface{}{}
		for _, i := range r.Images {
			sqlString += "(?, ?, ?, ?),"
			v = append(v, lid, i.URL, i.Alt, cid)
		}
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, _ = ds.db.Prepare(sqlString)
		defer stmt.Close()

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Printf("savePageReport\nCID: %v\n Images: %+v\nError: %+v\n", cid, v, err)
		}
	}

	if len(r.Iframes) > 0 {
		sqlString := "INSERT INTO iframes (pagereport_id, url, crawl_id) values "

		v := []interface{}{}
		for _, i := range r.Iframes {
			sqlString += "(?, ?, ?),"
			v = append(v, lid, i, cid)
		}
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, _ = ds.db.Prepare(sqlString)
		defer stmt.Close()

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Printf("savePageReport\nCID: %v\n Iframes: %+v\nError: %+v\n", cid, v, err)
		}
	}

	if len(r.Audios) > 0 {
		sqlString := "INSERT INTO audios (pagereport_id, url, crawl_id) values "

		v := []interface{}{}
		for _, i := range r.Audios {
			sqlString += "(?, ?, ?),"
			v = append(v, lid, i, cid)
		}
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, _ = ds.db.Prepare(sqlString)
		defer stmt.Close()

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Printf("savePageReport\nCID: %v\n Audios: %+v\nError: %+v\n", cid, v, err)
		}
	}

	if len(r.Videos) > 0 {
		sqlString := "INSERT INTO videos (pagereport_id, url, crawl_id) values "

		v := []interface{}{}
		for _, i := range r.Videos {
			sqlString += "(?, ?, ?),"
			v = append(v, lid, i, cid)
		}
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, _ = ds.db.Prepare(sqlString)
		defer stmt.Close()

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Printf("savePageReport\nCID: %v\n Videos: %+v\nError: %+v\n", cid, v, err)
		}
	}

	if len(r.Scripts) > 0 {
		sqlString := "INSERT INTO scripts (pagereport_id, url, crawl_id) values "
		v := []interface{}{}
		for _, s := range r.Scripts {
			sqlString += "(?, ?, ?),"
			v = append(v, lid, s, cid)
		}
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, _ := ds.db.Prepare(sqlString)
		defer stmt.Close()

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Printf("savePageReport\nCID: %v\n Scripts: %+v\nError: %+v\n", cid, v, err)
		}
	}

	if len(r.Styles) > 0 {
		sqlString := "INSERT INTO styles (pagereport_id, url, crawl_id) values "
		v := []interface{}{}

		for _, s := range r.Styles {
			sqlString += "(?, ?, ?),"
			v = append(v, lid, s, cid)

		}
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, _ := ds.db.Prepare(sqlString)
		defer stmt.Close()

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Printf("savePageReport\nCID: %v\n Styles: %+v\nError: %+v\n", cid, v, err)
		}
	}
}

func (ds *Datastore) FindAllPageReportsByCrawlId(cid int64) <-chan *crawler.PageReport {
	prStream := make(chan *crawler.PageReport)

	go func() {
		defer close(prStream)

		query := `
			SELECT
				id,
				url,
				redirect_url,
				refresh,
				status_code,
				content_type,
				media_type,
				lang,
				title,
				description,
				robots,
				noindex,
				canonical,
				h1,
				h2,
				words,
				size,
				valid_headings,
				robotstxt_blocked,
				crawled
			FROM pagereports
			WHERE crawl_id = ?`

		rows, err := ds.db.Query(query, cid)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			p := &crawler.PageReport{}
			err := rows.Scan(&p.Id,
				&p.URL,
				&p.RedirectURL,
				&p.Refresh,
				&p.StatusCode,
				&p.ContentType,
				&p.MediaType,
				&p.Lang,
				&p.Title,
				&p.Description,
				&p.Robots,
				&p.Noindex,
				&p.Canonical,
				&p.H1,
				&p.H2,
				&p.Words,
				&p.Size,
				&p.ValidHeadings,
				&p.BlockedByRobotstxt,
				&p.Crawled,
			)
			if err != nil {
				log.Println(err)
				continue
			}

			prStream <- p
		}
	}()

	return prStream
}

func (ds *Datastore) FindAllPageReportsByCrawlIdAndErrorType(cid int64, et string) <-chan *crawler.PageReport {
	prStream := make(chan *crawler.PageReport)

	go func() {
		defer close(prStream)

		query := `
			SELECT
				id,
				url,
				redirect_url,
				refresh,
				status_code,
				content_type,
				media_type,
				lang,
				title,
				description,
				robots,
				noindex,
				canonical,
				h1,
				h2,
				words,
				size,
				valid_headings,
				robotstxt_blocked,
				crawled
			FROM pagereports
			WHERE crawl_id = ?
			AND id IN (
				SELECT
					pagereport_id
				FROM issues
				INNER JOIN issue_types ON issue_types.id = issues.issue_type_id
				WHERE issue_types.type = ? AND crawl_id = ?
			)`

		rows, err := ds.db.Query(query, cid, et, cid)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			p := &crawler.PageReport{}
			err := rows.Scan(&p.Id,
				&p.URL,
				&p.RedirectURL,
				&p.Refresh,
				&p.StatusCode,
				&p.ContentType,
				&p.MediaType,
				&p.Lang,
				&p.Title,
				&p.Description,
				&p.Robots,
				&p.Noindex,
				&p.Canonical,
				&p.H1,
				&p.H2,
				&p.Words,
				&p.Size,
				&p.ValidHeadings,
				&p.BlockedByRobotstxt,
				&p.Crawled,
			)
			if err != nil {
				log.Println(err)
				continue
			}

			prStream <- p
		}
	}()

	return prStream
}

func (ds *Datastore) FindPageReportById(rid int) crawler.PageReport {
	query := `
		SELECT
			id,
			url,
			redirect_url,
			refresh,
			status_code,
			content_type,
			media_type,
			lang,
			title,
			description,
			robots,
			noindex,
			canonical,
			h1,
			h2,
			words,
			size,
			valid_headings,
			robotstxt_blocked,
			crawled
		FROM pagereports
		WHERE id = ?`

	row := ds.db.QueryRow(query, rid)

	p := crawler.PageReport{}
	err := row.Scan(&p.Id,
		&p.URL,
		&p.RedirectURL,
		&p.Refresh,
		&p.StatusCode,
		&p.ContentType,
		&p.MediaType,
		&p.Lang,
		&p.Title,
		&p.Description,
		&p.Robots,
		&p.Noindex,
		&p.Canonical,
		&p.H1,
		&p.H2,
		&p.Words,
		&p.Size,
		&p.ValidHeadings,
		&p.BlockedByRobotstxt,
		&p.Crawled,
	)
	if err != nil {
		log.Println(err)
	}

	hrows, err := ds.db.Query("SELECT to_url, to_lang FROM hreflangs WHERE pagereport_id = ?", rid)
	if err != nil {
		log.Println(err)
	}

	for hrows.Next() {
		h := crawler.Hreflang{}
		err = hrows.Scan(&h.URL, &h.Lang)
		if err != nil {
			log.Println(err)
			continue
		}

		p.Hreflangs = append(p.Hreflangs, h)
	}

	irows, err := ds.db.Query("SELECT url, alt FROM images WHERE pagereport_id = ?", rid)
	if err != nil {
		log.Println(err)
	}

	for irows.Next() {
		i := crawler.Image{}
		err = irows.Scan(&i.URL, &i.Alt)
		if err != nil {
			log.Println(err)
			continue
		}

		p.Images = append(p.Images, i)
	}

	ifrows, err := ds.db.Query("SELECT url FROM iframes WHERE pagereport_id = ?", rid)
	if err != nil {
		log.Println(err)
	}

	for ifrows.Next() {
		var url string
		err = ifrows.Scan(&url)
		if err != nil {
			log.Println(err)
			continue
		}

		p.Iframes = append(p.Iframes, url)
	}

	arows, err := ds.db.Query("SELECT url FROM audios WHERE pagereport_id = ?", rid)
	if err != nil {
		log.Println(err)
	}

	for arows.Next() {
		var url string
		err = arows.Scan(&url)
		if err != nil {
			log.Println(err)
			continue
		}

		p.Audios = append(p.Audios, url)
	}

	vrows, err := ds.db.Query("SELECT url FROM videos WHERE pagereport_id = ?", rid)
	if err != nil {
		log.Println(err)
	}

	for vrows.Next() {
		var url string
		err = vrows.Scan(&url)
		if err != nil {
			log.Println(err)
			continue
		}

		p.Videos = append(p.Videos, url)
	}

	scrows, err := ds.db.Query("SELECT url FROM scripts WHERE pagereport_id = ?", rid)
	if err != nil {
		log.Println(err)
	}

	for scrows.Next() {
		var url string
		err = scrows.Scan(&url)
		if err != nil {
			log.Println(err)
			continue
		}

		p.Scripts = append(p.Scripts, url)
	}

	strows, err := ds.db.Query("SELECT url FROM styles WHERE pagereport_id = ?", rid)
	if err != nil {
		log.Println(err)
	}

	for strows.Next() {
		var url string
		err = strows.Scan(&url)
		if err != nil {
			log.Println(err)
			continue
		}

		p.Styles = append(p.Styles, url)
	}

	return p
}

func (ds *Datastore) FindLinks(pageReport *crawler.PageReport, cid int64, p int) []crawler.Link {
	max := paginationMax
	offset := max * (p - 1)
	links := []crawler.Link{}

	query := `
		SELECT
			url,
			rel,
			nofollow,
			text
		FROM links
		WHERE pagereport_id = ?
		LIMIT ?,?
	`
	lrows, err := ds.db.Query(query, pageReport.Id, offset, max)
	if err != nil {
		log.Println(err)
	}

	for lrows.Next() {
		l := crawler.Link{}
		err = lrows.Scan(&l.URL, &l.Rel, &l.NoFollow, &l.Text)
		if err != nil {
			log.Println(err)
			continue
		}

		links = append(links, l)
	}

	return links
}

func (ds *Datastore) FindExternalLinks(pageReport *crawler.PageReport, cid int64, p int) []crawler.Link {
	max := paginationMax
	offset := max * (p - 1)
	links := []crawler.Link{}

	query := `
		SELECT
			url,
			rel,
			nofollow,
			text,
			Sponsored,
			UGC
		FROM external_links
		WHERE pagereport_id = ?
		LIMIT ?,?
	`

	lrows, err := ds.db.Query(query, pageReport.Id, offset, max)
	if err != nil {
		log.Println(err)
	}

	for lrows.Next() {
		l := crawler.Link{}
		err = lrows.Scan(&l.URL, &l.Rel, &l.NoFollow, &l.Text, &l.Sponsored, &l.UGC)
		if err != nil {
			log.Println(err)
			continue
		}

		links = append(links, l)
	}

	return links
}

func (ds *Datastore) FindSitemapPageReports(cid int64) <-chan *crawler.PageReport {
	prStream := make(chan *crawler.PageReport)

	go func() {
		defer close(prStream)

		query := `
			SELECT pagereports.id, pagereports.url, pagereports.title
			FROM pagereports
			WHERE media_type = "text/html" AND status_code >= 200 AND status_code < 300
			AND (canonical IS NULL OR canonical = "" OR canonical = url) AND pagereports.crawl_id = ?
			AND crawled = 1`

		rows, err := ds.db.Query(query, cid)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			p := &crawler.PageReport{}
			err := rows.Scan(&p.Id, &p.URL, &p.Title)
			if err != nil {
				log.Println(err)
				continue
			}

			prStream <- p
		}
	}()

	return prStream
}

func (ds *Datastore) FindPageReportIssues(cid int64, p int, errorType string) []crawler.PageReport {
	max := paginationMax
	offset := max * (p - 1)

	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE id IN (
			SELECT DISTINCT pagereport_id
			FROM issues
			INNER JOIN issue_types ON issue_types.id = issues.issue_type_id
			WHERE issue_types.type = ? AND crawl_id = ?
		) ORDER BY url ASC LIMIT ?, ?`

	var pageReports []crawler.PageReport
	rows, err := ds.db.Query(query, errorType, cid, offset, max)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		p := crawler.PageReport{}
		err := rows.Scan(&p.Id, &p.URL, &p.Title)
		if err != nil {
			log.Println(err)
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func (ds *Datastore) FindInLinks(s string, cid int64, p int) []crawler.PageReport {
	max := paginationMax
	offset := max * (p - 1)

	hash := Hash(s)
	query := `
		SELECT
			pagereports.id,
			pagereports.url,
			pagereports.Title
		FROM links
		LEFT JOIN pagereports ON pagereports.id = links.pagereport_id
		WHERE links.url_hash = ? AND pagereports.crawl_id = ? AND pagereports.crawled = 1
		GROUP BY pagereports.id
		LIMIT ?,?`

	var pageReports []crawler.PageReport
	rows, err := ds.db.Query(query, hash, cid, offset, max)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		p := crawler.PageReport{}
		err := rows.Scan(&p.Id, &p.URL, &p.Title)
		if err != nil {
			log.Println(err)
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func (ds *Datastore) FindPageReportsRedirectingToURL(u string, cid int64, p int) []crawler.PageReport {
	max := paginationMax
	offset := max * (p - 1)
	uh := Hash(u)
	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE redirect_hash = ? AND crawl_id = ? AND crawled = 1
		LIMIT ?,?`

	var pageReports []crawler.PageReport
	rows, err := ds.db.Query(query, uh, cid, offset, max)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		p := crawler.PageReport{}
		err := rows.Scan(&p.Id, &p.URL, &p.Title)
		if err != nil {
			log.Println(err)
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func (ds *Datastore) GetNumberOfPagesForLinks(pageReport *crawler.PageReport, cid int64) int {
	query := `
		SELECT
			count(*)
		FROM links
		WHERE pagereport_id = ? AND crawl_id = ?
	`

	row := ds.db.QueryRow(query, pageReport.Id, cid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("GetNumberOfPagesForLinks: %v\n", err)
	}
	var f float64 = float64(c) / float64(paginationMax)
	return int(math.Ceil(f))
}

func (ds *Datastore) GetNumberOfPagesForExternalLinks(pageReport *crawler.PageReport, cid int64) int {
	query := `
		SELECT
			count(*)
		FROM external_links
		WHERE pagereport_id = ? AND crawl_id = ?
	`

	row := ds.db.QueryRow(query, pageReport.Id, cid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("GetNumberOfPagesForExternalLinks: %v\n", err)
	}
	var f float64 = float64(c) / float64(paginationMax)
	return int(math.Ceil(f))
}

func (ds *Datastore) GetNumberOfPagesForInlinks(pageReport *crawler.PageReport, cid int64) int {
	h := Hash(pageReport.URL)
	query := `
		SELECT
			count(distinct pagereports.id)
		FROM links
		LEFT JOIN pagereports ON pagereports.id = links.pagereport_id
		WHERE links.url_hash = ? AND pagereports.crawl_id = ? AND pagereports.crawled = 1
	`

	row := ds.db.QueryRow(query, h, cid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("GetNumberOfPagesForInlinks: %v\n", err)
	}
	var f float64 = float64(c) / float64(paginationMax)
	return int(math.Ceil(f))
}

func (ds *Datastore) GetNumberOfPagesForRedirecting(pageReport *crawler.PageReport, cid int64) int {
	h := Hash(pageReport.URL)
	query := `
		SELECT
			count(id)
		FROM pagereports
		WHERE redirect_hash = ? AND crawl_id = ? AND crawled = 1
	`

	row := ds.db.QueryRow(query, h, cid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("GetNumberOfPagesForRedirecting: %v\n", err)
	}
	var f float64 = float64(c) / float64(paginationMax)
	return int(math.Ceil(f))

}

func (ds *Datastore) CountSponsoredLinks(cid int64) int {
	query := `
		SELECT
			count(id)
		FROM external_links
		WHERE sponsored = 1 AND crawl_id = ?
	`

	row := ds.db.QueryRow(query, cid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("CountSponsoredLinks: %v\n", err)
	}

	return c
}

func (ds *Datastore) CountUGCLinks(cid int64) int {
	query := `
		SELECT
			count(id)
		FROM external_links
		WHERE ugc = 1 AND crawl_id = ?
	`

	row := ds.db.QueryRow(query, cid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("CountUGCLinks: %v\n", err)
	}

	return c
}

func (ds *Datastore) CountByCanonical(cid int64) int {
	query := `
		SELECT
			count(id)
		FROM pagereports
		WHERE crawl_id = ? AND media_type = "text/html" AND (canonical = "" OR canonical = url)
			AND status_code >= 200 AND status_code < 300
	`

	row := ds.db.QueryRow(query, cid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("CountUGCLinks: %v\n", err)
	}

	return c
}

func (ds *Datastore) CountByNonCanonical(cid int64) int {
	query := `
		SELECT
			count(id)
		FROM pagereports
		WHERE crawl_id = ? AND media_type = "text/html" AND canonical != "" AND canonical != url
			AND status_code >= 200 AND status_code < 300
	`

	row := ds.db.QueryRow(query, cid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("CountUGCLinks: %v\n", err)
	}

	return c
}

func (ds *Datastore) CountImagesAlt(cid int64) *issue.AltCount {
	query := `
		SELECT
			if(alt = "", "no alt", "alt") as a,
			count(*)
		FROM images
		WHERE crawl_id = ?
		GROUP BY a
	`

	c := &issue.AltCount{}

	rows, err := ds.db.Query(query, cid)
	if err != nil {
		log.Println(err)
		return c
	}

	for rows.Next() {
		var v int
		var a string

		err := rows.Scan(&a, &v)
		if err != nil {
			continue
		}

		if a == "alt" {
			c.Alt = v
		} else {
			c.NonAlt = v
		}
	}

	return c
}

func (ds *Datastore) CountScheme(cid int64) *issue.SchemeCount {
	query := `
		SELECT
			scheme,
			count(*)
		FROM pagereports
		WHERE crawl_id = ?
		GROUP BY scheme
	`

	c := &issue.SchemeCount{}

	rows, err := ds.db.Query(query, cid)
	if err != nil {
		log.Println(err)
		return c
	}

	for rows.Next() {
		var v int
		var a string

		err := rows.Scan(&a, &v)
		if err != nil {
			continue
		}

		if a == "https" {
			c.HTTPS = v
		} else {
			c.HTTP = v
		}
	}

	return c
}
