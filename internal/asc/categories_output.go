package asc

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// formatPlatforms converts a slice of Platform to a comma-separated string.
func formatPlatforms(platforms []Platform) string {
	strs := make([]string, len(platforms))
	for i, p := range platforms {
		strs[i] = string(p)
	}
	return strings.Join(strs, ", ")
}

func printAppCategoriesMarkdown(resp *AppCategoriesResponse) error {
	fmt.Println("| ID | Platforms |")
	fmt.Println("|---|---|")
	for _, cat := range resp.Data {
		fmt.Printf("| %s | %s |\n", cat.ID, formatPlatforms(cat.Attributes.Platforms))
	}
	return nil
}

func printAppCategoriesTable(resp *AppCategoriesResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tPLATFORMS")
	for _, cat := range resp.Data {
		fmt.Fprintf(w, "%s\t%s\n", cat.ID, formatPlatforms(cat.Attributes.Platforms))
	}
	return w.Flush()
}
