package main

import (
	"html/template"
	"log"
	"time"
)

func unixToTime(unix int64) string {
	return time.Unix(unix, 0).Format(time.RFC3339)
}

func getTemplates() *template.Template {
	t, err := template.New("dump").Funcs(template.FuncMap{
		"unixToTime": unixToTime,
	}).Parse(`<!DOCTYPE html>
<html lang="en" charset="utf-8">
<head>
	<title>Stat server</title>
	<meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
</head>
<body>
{{ range $key, $value := . }}
  <h2>Server: {{ $key }}</h2>
	<h3>Hour data</h3>
  <table>
    <thead>
      <tr>
         <th>StartTime</th>
         <th>StartTimeUnix</th>
         <th>RecordCount</th>
         <th>AvgCPU</th>
         <th>AvgRAM</th>
      </tr>
    </thead>
    <tbody>
    {{ range $hd := $value.HourData }}
      <tr>
        <td>{{$hd.StartUnixTime | unixToTime}}</td>
        <td>{{$hd.StartUnixTime}}</td>
        <td>{{$hd.Count}}</td>
        <td>{{$hd.AvgCPU}}</td>
        <td>{{$hd.AvgRAM}}</td>
      </tr>
    {{ end }}
    </tbody>
  </table>
  <h3>Minute data</h3>
  <table>
    <thead>
      <tr>
         <th>StartTime</th>
         <th>StartTimeUnix</th>
         <th>RecordCount</th>
         <th>AvgCPU</th>
         <th>AvgRAM</th>
      </tr>
    </thead>
    <tbody>
    {{ range $md := $value.MinuteData }}
      <tr>
        <td>{{ $md.StartUnixTime | unixToTime }}</td>
        <td>{{$md.StartUnixTime}}</td>
        <td>{{$md.Count}}</td>
        <td>{{$md.AvgCPU}}</td>
        <td>{{$md.AvgRAM}}</td>
      </tr>
    {{ end }}
    </tbody>
  </table>
  <hr>
{{ end }}
</body>
</html>`)
	if err != nil {
		log.Fatalf("failed parsing template %s", err)
	}

	return t
}
