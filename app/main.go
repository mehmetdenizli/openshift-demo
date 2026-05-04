package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// HTML template for the demo page
var htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OpenShift Demo App</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: 'Segoe UI', sans-serif; background: #0f1117; color: #e0e0e0; min-height: 100vh; }
        .header { background: linear-gradient(135deg, #cc0000, #8b0000); padding: 24px 40px; }
        .header h1 { font-size: 2rem; color: #fff; }
        .header p  { color: #ffcccc; margin-top: 4px; }
        .container { max-width: 1000px; margin: 40px auto; padding: 0 24px; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 20px; }
        .card { background: #1a1d2e; border: 1px solid #2a2d3e; border-radius: 10px; padding: 24px; }
        .card h2 { font-size: 1rem; color: #888; text-transform: uppercase; letter-spacing: 1px; margin-bottom: 16px; }
        .badge { display: inline-block; background: #0d3349; border: 1px solid #1a6396; color: #5eb8ff;
                 border-radius: 4px; padding: 4px 10px; font-family: monospace; font-size: 0.85rem; margin: 3px 2px; }
        .badge.green  { background: #0d3d1f; border-color: #1a7a3f; color: #4dcd7f; }
        .badge.orange { background: #3d2600; border-color: #7a4d00; color: #ffaa33; }
        .badge.purple { background: #2d1a4d; border-color: #5a3496; color: #b388ff; }
        .value { font-family: monospace; color: #5eb8ff; word-break: break-all; }
        .footer { text-align: center; padding: 30px; color: #555; font-size: 0.85rem; }
        .hit  { font-size: 3rem; font-weight: bold; color: #cc0000; }
    </style>
</head>
<body>
<div class="header">
    <h1>🚀 OpenShift Demo App</h1>
    <p>Written in Go — Demonstrating OpenShift core concepts</p>
</div>
<div class="container">
    <div class="grid">

        <!-- Deployment Info -->
        <div class="card">
            <h2>📦 Deployment Info</h2>
            <p>Pod Name<br><span class="value">{{.PodName}}</span></p>
            <br>
            <p>Namespace<br><span class="value">{{.Namespace}}</span></p>
            <br>
            <p>Node<br><span class="value">{{.NodeName}}</span></p>
        </div>

        <!-- ConfigMap Values -->
        <div class="card">
            <h2>⚙️ ConfigMap Values</h2>
            <p>APP_ENV<br><span class="badge">{{.AppEnv}}</span></p>
            <br>
            <p>APP_COLOR<br><span class="badge orange">{{.AppColor}}</span></p>
            <br>
            <p>APP_MESSAGE<br><span class="value">{{.AppMessage}}</span></p>
        </div>

        <!-- Secret Values -->
        <div class="card">
            <h2>🔐 Secret Values</h2>
            <p>DB_USER<br><span class="badge green">{{.DbUser}}</span></p>
            <br>
            <p>DB_PASSWORD<br><span class="badge green">*** ({{.DbPassLen}} characters)</span></p>
            <br>
            <p>API_KEY<br><span class="badge green">*** (hidden)</span></p>
        </div>

        <!-- PV / PVC Storage -->
        <div class="card">
            <h2>💾 PVC / Storage</h2>
            <p>Mount Point<br><span class="value">/data/app</span></p>
            <br>
            <p>Written File<br><span class="value">{{.StorageFile}}</span></p>
            <br>
            <p>File Content<br><span class="badge purple">{{.StorageContent}}</span></p>
        </div>

        <!-- Network / Service -->
        <div class="card">
            <h2>🌐 Network / Service & Route</h2>
            <p>Service Port<br><span class="badge">8080/TCP</span></p>
            <br>
            <p>Route URL<br><span class="value">{{.RouteURL}}</span></p>
            <br>
            <p>Visit Count<br><span class="hit">{{.HitCount}}</span></p>
        </div>

        <!-- Image Info -->
        <div class="card" style="border-top: 4px solid #5eb8ff;">
            <h2>🖼️ Image & Build Info</h2>
            <p>Version / Tag<br><span class="badge purple" style="font-size: 1.2rem;">{{.AppVersion}}</span></p>
            <br>
            <p>Current Image<br><span class="value" style="color: #ffaa33;">{{.ImageRef}}</span></p>
            <br>
            <p>Build Date<br><span class="badge">{{.BuildDate}}</span></p>
            <br>
            <p>Go Version<br><span class="badge green">{{.GoVersion}}</span></p>
        </div>

    </div>
</div>
<div class="footer">OpenShift CRC Demo · {{.Timestamp}}</div>
</body>
</html>
`

// Global hit counter (in-memory, resets on pod restart — intentional for demo)
var hitCount int

// pageData holds all template variables
type pageData struct {
	PodName        string
	Namespace      string
	NodeName       string
	AppEnv         string
	AppColor       string
	AppMessage     string
	AppVersion     string
	DbUser         string
	DbPassLen      int
	StorageFile    string
	StorageContent string
	RouteURL       string
	HitCount       int
	ImageRef       string
	BuildDate      string
	GoVersion      string
	Timestamp      string
}

// env returns the value of the given env var, or fallback if not set.
func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// writeStorageFile creates a timestamped file in /data/app to demonstrate PVC usage.
func writeStorageFile() (filename, content string) {
	dir := "/data/app"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "error: directory could not be created", err.Error()
	}
	filename = fmt.Sprintf("%s/hit-%d.txt", dir, hitCount)
	content = fmt.Sprintf("visit=%d time=%s", hitCount, time.Now().Format(time.RFC3339))
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return filename, "could not write: " + err.Error() + " (path: " + filename + ")"
	}
	return filename, content
}

func handler(w http.ResponseWriter, r *http.Request) {
	hitCount++

	// Read env vars injected via ConfigMap
	appEnv     := env("APP_ENV", "development")
	appColor   := env("APP_COLOR", "red")
	appMessage := env("APP_MESSAGE", "Hello OpenShift!")
	appVersion := env("APP_VERSION", "v1.0.0")

	// Read env vars injected via Secret
	dbUser  := env("DB_USER", "(no secret)")
	dbPass  := env("DB_PASSWORD", "")
	apiKey  := env("API_KEY", "")
	_ = apiKey // shown as hidden in UI

	// PVC write demo
	storageFile, storageContent := writeStorageFile()
	// shorten path for display
	storageFile = strings.TrimPrefix(storageFile, "/data/app/")

	tmpl := template.Must(template.New("page").Parse(htmlTemplate))
	data := pageData{
		PodName:        env("POD_NAME", env("HOSTNAME", "unknown")),
		Namespace:      env("POD_NAMESPACE", "demo"),
		NodeName:       env("NODE_NAME", "unknown"),
		AppEnv:         appEnv,
		AppColor:       appColor,
		AppMessage:     appMessage,
		AppVersion:     appVersion,
		DbUser:         dbUser,
		DbPassLen:      len(dbPass),
		StorageFile:    storageFile,
		StorageContent: storageContent,
		RouteURL:       env("ROUTE_URL", "http://demo-app.apps-crc.testing"),
		HitCount:       hitCount,
		ImageRef:       env("IMAGE_REF", "image-registry.openshift-image-registry.svc:5000/demo/demo-app:latest"),
		BuildDate:      env("BUILD_DATE", "set during build"),
		GoVersion:      "go1.22",
		Timestamp:      time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
}

func main() {
	port := env("PORT", "8080")
	http.HandleFunc("/", handler)
	http.HandleFunc("/healthz", healthz)

	log.Printf("🚀 Demo app starting on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
