package templates

const topicindexSrc = `
{{ define "content" }}

<h1><a href="/groups?name={{ .GroupName }}">{{ .GroupName }}</a></h1>
<h2>{{ .TopicName }}</h2>

<div>
{{ .Content }}
</div>

{{ if or .IsAdmin .IsMod .IsSuperAdmin .IsOwner }}
<a href="/topics/edit?id={{ .TopicID }}">edit</a>
{{ end }}

{{ if not .IsClosed }}
<a href="/comments/new?tid={{ .TopicID }}">reply</a>
{{ end }}

{{ if .Comments }}
{{ range .Comments }}
<div class="row">
	<div>by {{ .UserName }} <a href="/comments?id={{ .ID }}">{{ .CreatedDate }}</a>{{ if or .IsOwner $.IsAdmin $.IsMod $.IsSuperAdmin }} | <a href="/comments/edit?id={{ .ID }}">edit</a> {{end}}</div>
	<div>{{ if .IsDeleted }}[DELETED]{{ else }}{{ .Content }}{{ end }}</div>
</div>
{{ end }}
{{ else }}
<div class="row">
	<div class="muted">No comments here.</div>
</div>
{{ end }}

{{ if .LastCommentDate }}
<div class="row">
	<div><a href="/topics?id={{ .TopicID }}&lcd={{ .LastCommentDate }}">More</a></div>
</div>
{{ end }}

{{ end }}`
