package cmd

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"queuectl.backend/internal/job"
	"queuectl.backend/internal/store"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start a simple web dashboard for monitoring the job queue",
	Run: func(cmd *cobra.Command, args []string) {
		CommonInit()
		db, err := store.InitDB()
		if err != nil {
			log.Fatalf("Failed to init DB: %v", err)
		}
		repo := store.NewJobRepo(db)

		tmpl := template.Must(template.New("dashboard").Parse(htmlTemplate))

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			stats, _ := repo.JobMetrics()
			jobs, _ := repo.ListJobs(nil, 100, 0, true)

			data := struct {
				Metrics store.MetricsSummary
				Jobs    []job.Job
			}{stats, jobs}

			w.Header().Set("Content-Type", "text/html")
			tmpl.Execute(w, data)
		})

		fmt.Println("✅ Web dashboard running at: http://localhost:8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	},
}

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>QueueCTL Dashboard</title>
<style>
  body { font-family: Arial, sans-serif; background: #f8fafc; margin: 0; padding: 20px; }
  h1 { text-align: center; color: #333; }
  table { width: 100%; border-collapse: collapse; margin-top: 20px; }
  th, td { padding: 8px 12px; border: 1px solid #ccc; text-align: left; }
  th { background: #333; color: #fff; }
  tr:nth-child(even) { background: #f2f2f2; }
  .state-pending { color: orange; font-weight: bold; }
  .state-processing { color: blue; font-weight: bold; }
  .state-completed { color: green; font-weight: bold; }
  .state-failed, .state-dead { color: red; font-weight: bold; }
  .stats { display: flex; justify-content: space-around; margin-top: 20px; background: #fff; padding: 10px; border-radius: 6px; box-shadow: 0 0 5px rgba(0,0,0,0.1); }
</style>
</head>
<body>
<h1>QueueCTL Dashboard</h1>

<div class="stats">
  <div><b>Total:</b> {{.Metrics.Total}}</div>
  <div><b>Pending:</b> {{.Metrics.Pending}}</div>
  <div><b>Processing:</b> {{.Metrics.Processing}}</div>
  <div><b>Completed:</b> {{.Metrics.Completed}}</div>
  <div><b>Failed:</b> {{.Metrics.Failed}}</div>
  <div><b>DLQ:</b> {{.Metrics.Dead}}</div>
</div>

<table>
  <tr>
    <th>ID</th>
    <th>Command</th>
    <th>State</th>
    <th>Priority</th>
    <th>Attempts</th>
    <th>Run At</th>
    <th>Duration (s)</th>
  </tr>
  {{range .Jobs}}
  <tr>
    <td>{{.ID}}</td>
    <td>{{.Command}}</td>
    <td class="state-{{.State}}">{{.State}}</td>
    <td>{{.Priority}}</td>
    <td>{{.Attempts}} / {{.MaxRetries}}</td>
    <td>{{if .RunAt}}{{.RunAt.Format "15:04:05"}}{{else}}-{{end}}</td>
    <td>{{printf "%.2f" .Duration}}</td>
  </tr>
  {{end}}
</table>
</body>
</html>
`

func init() {
	rootCmd.AddCommand(webCmd)
}
