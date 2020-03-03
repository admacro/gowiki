# Gowiki

Writing web application in Go, the [Go Wiki](https://golang.org/doc/articles/wiki/) tutorial.

### Implemented [Other tasks](https://golang.org/doc/articles/wiki/#tmp_14)
 - Store templates in `tmpl/` and page data in `data/`. [Done]
 - Add a handler to make the web root redirect to `/view/FrontPage`.
 - Spruce up the page templates by making them valid HTML and adding some CSS rules.
 - Implement inter-page linking by converting instances of [PageName] to
`<a href="/view/PageName">PageName</a>`. (hint: you could use regexp.ReplaceAllFunc to do this)
