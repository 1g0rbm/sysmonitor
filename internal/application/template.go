package application

const AllMetricsTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>All metrics</title>
</head>
<body>
<h1>List of metrics</h1>
<ul>
    {{range .}}
        <li>{{.Name}}:{{.ValueAsString}}</li>
    {{end}}
</ul>
</body>
</html>
`
