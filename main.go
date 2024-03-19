package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/denisenkom/go-mssqldb"
	"log"
	"os"
	"time"

	"path/filepath"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table table.Model
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
			// case "enter":
			// 	return m, tea.Batch(
			// 		tea.Printf("Let's go to %s!", m.table.SelectedRow()[1]),
			// 	)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func tables_used(db *sql.DB, ctx context.Context) {
	tsql := `
		    
SELECT 
    OBJECT_NAME(s.object_id) AS TableName,
    SUM(user_seeks + user_scans + user_lookups) AS TotalAccesses,
    MAX(last_access_date) AS LastAccessDate
FROM 
    sys.dm_db_index_usage_stats AS s
INNER JOIN 
    sys.objects AS o ON s.object_id = o.object_id
CROSS APPLY (
    VALUES 
    (s.last_user_seek),
    (s.last_user_scan),
    (s.last_user_lookup)
) AS last_access(last_access_date)
WHERE 
    OBJECTPROPERTY(s.object_id,'IsUserTable') = 1
GROUP BY 
    OBJECT_NAME(s.object_id)
ORDER BY 
    SUM(user_seeks + user_scans + user_lookups) DESC

		`

	// Execute query
	rows, err := db.QueryContext(ctx, tsql)
	if err != nil {
		log.Fatal("Error executing query: ", err.Error())
	}
	defer rows.Close()

	var columns []table.Column = []table.Column{
		{Title: "Table", Width: 50},
		{Title: "Count", Width: 5},
		{Title: "Last Accessed", Width: 20},
	}

	var bubbletea_rows []table.Row = []table.Row{}
	// Iterate through the result set
	for rows.Next() {
		var tableName string
		var totalAccesses int
		var lastAccessDate time.Time
		err := rows.Scan(&tableName, &totalAccesses, &lastAccessDate)
		if err != nil {
			log.Fatal("Error scanning row: ", err.Error())
		}

		lastAccessDateString := lastAccessDate.Format("2006-01-02")
		var new_row table.Row = table.Row{tableName, fmt.Sprintf("%d", totalAccesses), lastAccessDateString}
		bubbletea_rows = append(bubbletea_rows, new_row)
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(bubbletea_rows),
		table.WithFocused(true),
		table.WithHeight(25),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := model{t}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

type Config struct {
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Server   string `json:"server"`
	Database string `json:"database"`
}

func main() {
	// Define flags
	newConnection := flag.Bool("new-connection", false, "Create a new connection")
	nc := flag.Bool("nc", false, "Create a new connection")

	// Parse command-line arguments
	flag.Parse()

	if *newConnection || *nc {
		// Get the user's home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error getting user home directory: %s", err)
		}

		// Construct the full path, replacing '~' with the user's home directory
		configPath := filepath.Join(homeDir, ".config/microscope.json")

		// Load the configuration file using os.ReadFile
		configBytes, err := os.ReadFile(configPath)
		if err != nil {
			log.Fatalf("Error reading config file: %s", err)
		}

		// Parse the JSON configuration
		var config Config
		if err := json.Unmarshal(configBytes, &config); err != nil {
			log.Fatalf("Error parsing config JSON: %s", err)
		}

		// Create connection pool
		connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;",
			config.Server, config.User, config.Password, config.Port, config.Database)
		db, err := sql.Open("sqlserver", connString)
		if err != nil {
			log.Fatalf("Error creating connection pool: %s", err)
		}

		ctx := context.Background()
		if err = db.PingContext(ctx); err != nil {
			log.Fatal(err.Error())
		}

		fmt.Println("Connected!")

		tables_used(db, ctx)

	} else {
		fmt.Println("Usage: go run main.go -new-connection")
		os.Exit(1)
	}
}
