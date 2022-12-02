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

    .topright {
        position: absolute;
        top: 0px;
        right: 0px;
        padding: 5px;
    }
  </style>
</head>
<body>
<div class="topright">
    Missing logs? Scan your own:<br>
    <a href="{{.Oauth2LoginUrl}}" target="_blank">Log into Warcraft Logs Account</a>
</div>
{{- template "body" .}}
</body>
</html>
{{end}}`
