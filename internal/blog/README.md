# Blog module

A simple blog module for darkstorm-backend.

## Requests

### Author info

#### Get author info

> GET /author/{authorID}

```json
{
  id: "authorID",
  name: "author name",
  about: "about",
  picurl: "picture URL"
}
```

#### Update author info

> POST /author/{authorID}

Must have a auth token for a user with the `"blog": "admin"` permission.

```json
{
  name: "author name",
  about: "about",
  picurl: "picture url"
}
```

#### Add Author info

> POST /author

Must have a auth token for a user with the `"blog": "admin"` permission.

```json
{
  name: "author name",
  about: "about",
  picurl: "picture URL"
}
```

### Blog

#### Specific blog

> GET /blog/{blogID}

Return:

```json
{
  id: "blogID",
  staticPage: false, // static pages don't show up alongside other blog pages.
  createTime: 0, // creation time in Unix format
  updateTime: 0, // last update time in Unix format
  author: "authorID",
  favicon: "favicon url",
  title: "blog title",
  blog: "blog", // blog will have been converted to HTML
}
```

#### Create blog

Request:

> POST /blog

Must have a auth token for a user with the `"blog": "admin"` permission.

```json
{
  favicon: "favicon url",
  title: "blog title",
  blog: "blog", // blog will have been converted to HTML
}
```

Return:

```json
{
  id: "blogID"
}
```

#### Update blog

Request:

> POST /blog/{blogID}

Must have a auth token for a user with the `"blog": "admin"` permission.

```json
{
  favicon: "new icon",
  title: "new title",
  blog: "new blog content"
}
```

#### Latest blogs

> GET /blog?page=0

Will return up to 5 blogs. `page` query is optional (implies 0 if not set).

Return:

```json
{
  num: 1, // Number of returned results, returns up to 5 results
  blogs: [
    {
      id: "blogID",
      createTime: 0, // creation time in Unix format
      updateTime: 0, // last update time in Unix format
      author: "authorID",
      favicon: "favicon url",
      title: "blog title",
      blog: "blog", // blog will have been converted to HTML
    }
    ...
  ]
}
```

#### Blog List

> GET /blog/list?page=0

Will return up to 50 IDs. `page` query is optional (implies 0 if not set).

Return:

```json
{
  num: 1, // Number of returned results, returns up to 50 results
  blogList: [
    {
      id: "blogID",
      createTime: 0, // Unix format
    },
    {
      id: "blogID",
      createTime: 0, // Unix format
    },
    ...
  ]
}
```

### Portfolio

#### Get Projects

> GET: /portfolio?lang=go

Return:

```json
[
  {
    title: "Darkstorm Server",
    order: 0,
    repository: "https://github.com/CalebQ42/darkstorm-server",
    description: "The backend that runs runs my website and APIs",
    technologies: [ // May be empty
      "MongoDB",
      "RESTful API"
    ],
    language: [
      {
        language: "go",
        dates: "September 2021"
      }
    ]
  }
]
```
