package cmd

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

var uiPort string

var uiCmd = &cobra.Command{
	Use:   "ui [path]",
	Short: "Launch the interactive local security dashboard",
	Long:  `Scan the project and open a modern web-based dashboard to visualize findings and insights.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runUI,
}

func init() {
	rootCmd.AddCommand(uiCmd)
	uiCmd.Flags().StringVarP(&uiPort, "port", "p", "8080", "port to run the dashboard on")
}

func runUI(cmd *cobra.Command, args []string) {
	scanPath := "."
	if len(args) > 0 {
		scanPath = args[0]
	}

	fmt.Printf("%s🛡️ Tharos is preparing your dashboard...%s\n", colorCyan, colorReset)
	start := time.Now()
	results := analyzePath(scanPath, aiEnabled)
	duration := time.Since(start)

	totalVulns := 0
	for _, r := range results {
		totalVulns += len(r.Findings)
	}

	output := BatchResult{
		Results: results,
		Summary: ScanSummary{
			TotalFiles:      len(results),
			Vulnerabilities: totalVulns,
			Duration:        duration.String(),
			DurationMs:      duration.Milliseconds(),
		},
	}

	// Prepare the dashboard HTML (using a template for potential dynamic data)
	tmpl, err := template.New("dashboard").Delims("[[", "]]").Parse(dashboardTemplate)
	if err != nil {
		fmt.Printf("❌ Failed to parse dashboard template: %v\n", err)
		return
	}

	// Handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/api/results", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(output)
	})

	url := fmt.Sprintf("http://localhost:%s", uiPort)
	fmt.Printf("\n%s🚀 Dashboard ready at: %s%s\n", colorGreen, url, colorReset)
	fmt.Printf("%s(Press Ctrl+C to stop)%s\n", colorGray, colorReset)

	// Auto-open browser
	go openBrowser(url)

	if err := http.ListenAndServe(":"+uiPort, nil); err != nil {
		fmt.Printf("❌ Failed to start dashboard server: %v\n", err)
	}
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start() // tharos-security-ignore
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start() // tharos-security-ignore
	case "darwin":
		err = exec.Command("open", url).Start() // tharos-security-ignore
	}
	if err != nil {
		fmt.Printf("⚠️  Could not auto-open browser: %v\n", err)
	}
}

const dashboardTemplate = `
<!DOCTYPE html>
<html lang="en" class="dark">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Tharos | Security Dashboard</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700&family=JetBrains+Mono&display=swap" rel="stylesheet">
    <script src="https://unpkg.com/lucide@latest"></script>
    <style>
        body { font-family: 'Inter', sans-serif; background-color: #030712; color: #f3f4f6; }
        .mono { font-family: 'JetBrains Mono', monospace; }
        .glass { background: rgba(17, 24, 39, 0.7); backdrop-filter: blur(12px); border: 1px solid rgba(255, 255, 255, 0.1); }
        .severity-critical { color: #f87171; border-left: 4px solid #f87171; }
        .severity-high { color: #fb923c; border-left: 4px solid #fb923c; }
        .severity-medium { color: #facc15; border-left: 4px solid #facc15; }
        .severity-info { color: #60a5fa; border-left: 4px solid #60a5fa; }
    </style>
</head>
<body class="p-6">
    <div id="app" class="max-w-6xl mx-auto">
        <!-- Header -->
        <header class="flex justify-between items-center mb-10">
            <div class="flex items-center gap-3">
                <div class="w-10 h-10 bg-gradient-to-br from-orange-500 to-red-600 rounded-lg flex items-center justify-center shadow-lg">
                    <i data-lucide="shield-check" class="text-white w-6 h-6"></i>
                </div>
                <div>
                    <h1 class="text-2xl font-black italic tracking-tighter">THAROS <span class="text-orange-500">DASHBOARD</span></h1>
                    <p class="text-slate-400 text-[10px] uppercase font-bold tracking-widest opacity-50">Local Security Analysis Engine</p>
                </div>
            </div>
            <div id="summary-badges" class="flex gap-4">
                <!-- Badges injected here -->
            </div>
        </header>

        <!-- Stats Grid -->
        <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-10">
            <div class="glass p-5 rounded-xl">
                <p class="text-slate-400 text-sm mb-1">Total Files</p>
                <p id="stat-files" class="text-3xl font-bold">--</p>
            </div>
            <div class="glass p-5 rounded-xl">
                <p class="text-slate-400 text-sm mb-1">Vulnerabilities</p>
                <p id="stat-findings" class="text-3xl font-bold text-orange-500">--</p>
            </div>
            <div class="glass p-5 rounded-xl md:col-span-2 flex items-center justify-between">
                <div>
                    <p class="text-slate-400 text-sm mb-1">Scan Duration</p>
                    <p id="stat-duration" class="text-3xl font-bold">--</p>
                </div>
                <div class="text-right">
                    <button onclick="window.location.reload()" class="bg-orange-600 hover:bg-orange-500 text-white px-4 py-2 rounded-lg text-sm font-semibold transition-all flex items-center gap-2">
                        <i data-lucide="refresh-cw" class="w-4 h-4"></i> Rescan
                    </button>
                </div>
            </div>
        </div>

        <!-- Main Content -->
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
            <!-- Sidebar: File List -->
            <div class="lg:col-span-1 glass rounded-xl overflow-hidden flex flex-col max-h-[70vh]">
                <div class="p-4 border-b border-white/10 bg-white/5 font-semibold flex justify-between">
                    Analyzed Files
                    <span id="file-count" class="text-xs bg-white/10 px-2 py-0.5 rounded">0</span>
                </div>
                <div id="file-list" class="overflow-y-auto flex-1">
                    <!-- Files injected here -->
                </div>
            </div>

            <!-- Details: Findings -->
            <div class="lg:col-span-2 space-y-4 overflow-y-auto max-h-[70vh]" id="findings-container">
                <div class="flex flex-col items-center justify-center h-full text-slate-500">
                    <i data-lucide="search" class="w-12 h-12 mb-4 opacity-20"></i>
                    <p>Select a file to see security findings</p>
                </div>
            </div>
        </div>
    </div>

    <!-- Templates -->
    <template id="finding-card">
        <div class="glass p-5 rounded-xl border-l-4">
            <div class="flex justify-between items-start mb-3">
                <h3 class="font-bold text-lg finding-rule">--</h3>
                <span class="text-xs font-bold uppercase py-1 px-2 rounded-md finding-severity-badge">--</span>
            </div>
            <p class="text-slate-300 mb-4 finding-message">--</p>
            
            <div class="space-y-3">
                <div class="bg-black/40 rounded-lg p-3">
                    <p class="text-xs text-slate-500 uppercase font-bold mb-1">Analysis</p>
                    <p class="text-sm finding-explain">--</p>
                </div>
                <div class="flex justify-between items-center text-xs">
                    <span class="text-slate-500">LINE <span class="finding-line">--</span></span>
                    <button class="text-orange-500 font-bold hover:underline">Magic Fix ✨</button>
                </div>
            </div>
        </div>
    </template>

    <script>
        let scanData = null;

        async function init() {
            try {
                const res = await fetch('/api/results');
                scanData = await res.json();
                renderDashboard();
                lucide.createIcons();
            } catch (err) {
                console.error("Failed to load results:", err);
            }
        }

        function renderDashboard() {
            if (!scanData) return;

            // Stats
            document.getElementById('stat-files').innerText = scanData.summary.total_files;
            document.getElementById('stat-findings').innerText = scanData.summary.vulnerabilities;
            document.getElementById('stat-duration').innerText = scanData.summary.duration;
            document.getElementById('file-count').innerText = scanData.summary.total_files;

            // File List
            const fileList = document.getElementById('file-list');
            fileList.innerHTML = '';
            
            scanData.results.forEach((result, idx) => {
                const div = document.createElement('div');
                const hasVulns = result.findings.length > 0;
                div.className = "p-4 border-b border-white/5 hover:bg-white/5 cursor-pointer transition-colors flex justify-between items-center " + (hasVulns ? "text-orange-400" : "text-slate-400");
                div.onclick = () => showFindings(idx);
                
                div.innerHTML = '' +
                    '<div class="truncate mr-4">' +
                        '<p class="text-sm font-medium truncate">' + result.file + '</p>' +
                    '</div>' +
                    (hasVulns ? ' <span class="bg-orange-500/20 text-orange-500 text-[10px] font-bold px-1.5 py-0.5 rounded">' + result.findings.length + '</span>' : '<i data-lucide="check-circle-2" class="w-4 h-4 text-emerald-500"></i>');
                fileList.appendChild(div);
            });

            // Initial view if any findings
            const firstWithVulns = scanData.results.findIndex(r => r.findings.length > 0);
            if (firstWithVulns >= 0) showFindings(firstWithVulns);
        }

        function showFindings(index) {
            const container = document.getElementById('findings-container');
            const result = scanData.results[index];
            container.innerHTML = '' +
                '<div class="mb-6">' +
                    '<h2 class="text-xl font-bold truncate">' + result.file + '</h2>' +
                    '<p class="text-slate-500 text-sm">Found ' + result.findings.length + ' issues in this file</p>' +
                '</div>';

            if (result.findings.length === 0) {
                container.innerHTML += '<div class="glass p-10 rounded-xl text-center text-emerald-500 font-bold border-l-4 border-emerald-500">✨ No vulnerabilities found in this file.</div>';
                lucide.createIcons();
                return;
            }

            const template = document.getElementById('finding-card');
            result.findings.forEach(f => {
                const clone = template.content.cloneNode(true);
                const sevClass = 'severity-' + (f.severity || 'info').toLowerCase();
                
                const card = clone.querySelector('.glass');
                card.classList.add(sevClass);

                clone.querySelector('.finding-rule').innerText = f.rule;
                clone.querySelector('.finding-message').innerText = f.message;
                clone.querySelector('.finding-explain').innerText = f.explain;
                clone.querySelector('.finding-line').innerText = f.line;

                const badge = clone.querySelector('.finding-severity-badge');
                badge.innerText = f.severity;
                badge.className += ' ' + getBadgeColors(f.severity);

                container.appendChild(clone);
            });
            
            lucide.createIcons();
        }

        function getBadgeColors(sev) {
            switch((sev || '').toLowerCase()) {
                case 'critical': return 'bg-red-500/20 text-red-500';
                case 'high': return 'bg-orange-500/20 text-orange-500';
                case 'medium': return 'bg-yellow-500/20 text-yellow-500';
                default: return 'bg-blue-500/20 text-blue-500';
            }
        }

        init();
    </script>
</body>
</html>
`
