# Blog module

A simple blog module for darkstorm-backend.

## Requests

### Author info

> GET /author/{authorID}

```json
```

### Blog

#### Specific blog

> GET /blog/{blogID}

Return:

```json
{
    id: "blogID",
    createTime: 0, // creation time in Unix format
    updateTime: 0, // last update time in Unix format
    author: "authorID",
    favicon: "favicon url",
    title: "blog title",
    blog: "blog", // blog will have been converted to HTML
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
