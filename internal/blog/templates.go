package blog

const editor = `
<p id="editSelector" hx-target="#editPage" hx-swap="swap:0.5s settle:0.5s" hx-push-url="true">
	<a href="/editor/blog" hx-get="/editor/blog" class="editSelectorItems{{if eq .SelectedPage "blogs"}} editSelectorSelected{{end}}" {{if eq .SelectedPage "blogs"}}{{end}}>Blogs</a>
	<a href="/editor/portfolio" hx-get="/editor/portfolio" class="editSelectorItems{{if eq .SelectedPage "portfolio"}} editSelectorSelected{{end}}">Portfolio</a>
	<a href="/editor/author" hx-get="/editor/author" class="editSelectorItems{{if eq .SelectedPage "author"}} editSelectorSelected{{end}}">Author</a>
</p>
<div id="editPage">{{ .Page }}</div>`

type editorStruct struct {
	SelectedPage string
	Page         string
}

const blogPage = `<p>
	<label for="blog" style="margin-right:10px">Blog:</label>
	<select id="blogSelect"
			name='blog'
			hx-get='/editor/blog/edit'
			hx-target='#editor'>
		<option value=''{{if eq .Selected ""}} selected{{end}}>New Blog</option>
	{{ range $blog := .Blogs }}
		<option value='{{.ID}}'{{if eq $.Selected .ID}} selected{{end}}>{{.Title}}{{if .Draft}} (Draft){{end}}</option>
	{{end}}
	</select>
</p>
<div id="editor" hx-on::after-settle="editorResize()">{{.Editor}}</div>`

type blogPageStruct struct {
	Selected string
	Editor   string
	Blogs    []BlogList
}

const blogForm = `
<form id="editorForm" hx-post="/editor/blog/post" hx-target="#formResult" hx-confirm="Save changes, overwritting previous values??">
	<input name="id" type="hidden" value="{{.Blog.ID}}"></input>
	<p>
		<label for="staticPage" style="margin-right:10px">Static Page:</label><input type="checkbox" name="staticPage"{{if .Blog.StaticPage}} checked {{end}}/>
		<span class="vertical-seperator"></span>
		<label for="draft" style="margin-right:10px">Draft:</label><input type="checkbox" name="draft"{{if or .Blog.Draft (not .Blog.ID)}} checked {{end}}/>
	</p>
	<label for="title">Title</label>
	<input id="titleInput" name="title" value="{{.Blog.Title}}" type="text" onkeydown="return event.key != 'Enter';"/>
	<textarea id="textEditor" name="blog" oninput="editorResize()">{{.Blog.RawBlog}}</textarea>
	<div id="formResult">{{.Result}}</div>
	<p style="margin-right:0px;">
		<button class="formButton" type="submit">{{if eq .Blog.ID ""}}Create{{else}}Update{{end}}</button>
		<button class="formButton"
				hx-get="/editor/blog/edit"
				hx-include="#blogSelect"
				hx-target="#editor"
				hx-confirm="Undo all your changes??">
			Cancel
		</button>
	<p>
</form>`

type blogFormStruct struct {
	Blog   Blog
	Result string
}

const portfolioPage = `<p>
	<label for="project" style="margin-right:10px">Project:</label>
	<select id="projectSelect"
			name='project'
			hx-get='/editor/portfolio/edit'
			hx-target='#editor'>
		<option value=''{{if eq .Selected ""}} selected{{end}}>New Project</option>
	{{ range $project := .Projects }}
		<option value='{{.Title}}'{{if eq $.Selected .Title}} selected{{end}}>{{.Title}}</option>
	{{end}}
	</select>
</p>
<div id="editor" hx-on::after-settle="editorResize()">{{.Editor}}</div>`

type portfolioPageStruct struct {
	Selected string
	Editor   string
	Projects []PortfolioProject
}

// TODO: Add Languages to editor
const portfolioForm = `<form id="editorForm" hx-post="/editor/portfolio/post" hx-target="#formResult" hx-confirm="Save changes, overwritting previous values??">
	<input name="origTitle" type="hidden" value="{{.Project.Title}}"></input>
	<label for="title">Title</label>
	<input id="titleInput" name="title" value="{{.Project.Title}}" type="text" onkeydown="return event.key != 'Enter';"/>
	<label for="technologies">Technologies</label>
	<input id="techInput" name="technologies" value="{{range $ind, $tech := .Project.Technologies}}{{if not (eq $ind 0)}}, {{end}}$tech{{end}}" type="text" onkeydown="return event.key != 'Enter';"/>
	<textarea id="textEditor" name="description" oninput="editorResize()">{{.Project.Description}}</textarea>
	<div id="formResult">{{.Result}}</div>
	<p style="margin-right:0px;">
		<button class="formButton" type="submit">{{if eq .Project.Title ""}}Create{{else}}Update{{end}}</button>
		<button class="formButton"
				hx-get="/editor/portfolio/edit"
				hx-include="#projectSelect"
				hx-target="#editor"
				hx-confirm="Undo all your changes??">
			Cancel
		</button>
	<p>
</form>`

type portfolioFormStruct struct {
	Project PortfolioProject
	Result  string
}

const authorPage = `<p>
	<label for="author" style="margin-right:10px">Author:</label>
	<select id="authorSelect"
			name='author'
			hx-get='/editor/author/edit'
			hx-target='#editor'>
		<option value=''{{if eq .Selected ""}} selected{{end}}>New Author</option>
	{{ range $author := .Authors }}
		<option value='{{.ID}}'{{if eq $.Selected .ID}} selected{{end}}>{{.Name}}</option>
	{{end}}
	</select>
</p>
<div id="editor" hx-on::after-settle="editorResize()">{{.Editor}}</div>`

type authorPageStruct struct {
	Selected string
	Editor   string
	Authors  []Author
}

const authorForm = `<form id="editorForm" hx-post="/editor/author/post" hx-target="#formResult" hx-confirm="Save changes, overwritting previous values??">
	<input name="id" type="hidden" value="{{.Author.ID}}"></input>
	<label for="name">Name</label>
	<input id="nameInput" name="name" value="{{.Author.Name}}" type="text" onkeydown="return event.key != 'Enter';"/>
	<textarea id="textEditor" name="about" oninput="editorResize()">{{.Author.About}}</textarea>
	<div id="formResult">{{.Result}}</div>
	<p style="margin-right:0px;">
		<button class="formButton" type="submit">{{if eq .Author.ID ""}}Create{{else}}Update{{end}}</button>
		<button class="formButton"
				hx-get="/editor/author/edit"
				hx-include="#projectSelect"
				hx-target="#editor"
				hx-confirm="Undo all your changes??">
			Cancel
		</button>
	<p>
</form>`

type authorFormStruct struct {
	Author Author
	Result string
}
