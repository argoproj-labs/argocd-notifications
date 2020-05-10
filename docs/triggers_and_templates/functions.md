## Functions

Both templates and triggers have access to the set of functions.

Trigger example:

```yaml
name: app-operation-stuck
condition: time.Now().Sub(time.Parse(app.status.operationState.startedAt)).Minutes() >= 5
template: my-template
```

Template example:
```yaml
name: my-template
title: Application {{.app.metadata.name}} sync status is {{.app.status.sync.status}}
body: "Author: {{(call .repo.GetCommitMetadata .app.status.sync.revision).Author}}"
```

### **time**
Time related functions.

<hr>
**`time.Now() Time`**

Executes function built-in Golang [time.Now](https://golang.org/pkg/time/#Now) function.
Returns an instance of Golang [Time](https://golang.org/pkg/time/#Time).

<hr>
**`time.Parse(val string) Time`**

Parses specified string using RFC3339 layout. Returns an instance of Golang [Time](https://golang.org/pkg/time/#Time).

### **repo**
Functions that provide additional information about Application source repository.
<hr>
**`repo.RepoURLToHTTPS(url string) string`**

Transforms given GIT URL into HTTPs format.

<hr>
**`repo.FullNameByRepoURL(url string) string`**

Returns repository URL full name `(<owner>/<repoName>)`. Currently supports only Github, Gitlab and Bitbucket.

<hr>
**`repo.GetCommitMetadata(sha string) CommitMetadata`**

Returns commit metadata. The commit must belong to the application source repository. `CommitMetadata` fields:

* `Message string` commit message
* `Author string` - commit author
* `Date time.Time` - commit creation date  
* `Tags []string` - Associated tags
