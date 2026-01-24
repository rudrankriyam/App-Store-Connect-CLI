package asc

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

func printAppCategoriesMarkdown(resp *AppCategoriesResponse) error {
	fmt.Println("| ID | Platforms |")
	fmt.Println("|---|---|")
	for _, cat := range resp.Data {
		platforms := make([]string, len(cat.Attributes.Platforms))
		for i, p := range cat.Attributes.Platforms {
			platforms[i] = string(p)
		}
		fmt.Printf("| %s | %s |\n", cat.ID, strings.Join(platforms, ", "))
	}
	return nil
}

func printAppCategoriesTable(resp *AppCategoriesResponse) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tPLATFORMS")
	for _, cat := range resp.Data {
		platforms := make([]string, len(cat.Attributes.Platforms))
		for i, p := range cat.Attributes.Platforms {
			platforms[i] = string(p)
		}
		fmt.Fprintf(w, "%s\t%s\n", cat.ID, strings.Join(platforms, ", "))
	}
	return w.Flush()
}
