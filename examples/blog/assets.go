package blog

import _ "embed"

//go:embed templates/blog/post_detail.html
var postDetailTemplate string

//go:embed templates/blog/comment_form.html
var commentFormTemplate string

//go:embed static/blog/app.css
var blogStylesheet string

func Templates() map[string]string {
	return map[string]string{
		"blog/post_detail.html":  postDetailTemplate,
		"blog/comment_form.html": commentFormTemplate,
	}
}

func StaticFiles() map[string]string {
	return map[string]string{
		"blog/app.css": blogStylesheet,
	}
}
