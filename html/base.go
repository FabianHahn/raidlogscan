package html

const baseHtmlTemplate = `{{define "base"}}
<html>
<head>
  <title>{{.Title}} - WoW Raid Stats</title>
  <style type="text/css">
    a, a:visited, a:hover, a:active {
        color: inherit;
    }
    
    table {
        border-collapse: collapse;
        border: 1px solid black;
    }
    
    th {
        border: 1px solid black;
        padding: 3px;
    }
    
    td {
        border: 1px solid black;
        padding: 3px;
    }
    
    div {
        margin: 10px;
    }
    
    .column {
        float: left;
    }
  </style>
</head>
<body>
{{- template "body" .}}
</body>
</html>
{{end}}`
