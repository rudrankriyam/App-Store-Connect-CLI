package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
)

func printMigrateImportResultMarkdown(result *MigrateImportResult) error {
	if result.DryRun {
		fmt.Println("## Dry Run - No changes made")
		fmt.Println()
	}
	fmt.Printf("**Version ID:** %s\n\n", result.VersionID)

	// Version localizations
	fmt.Println("### Version Localizations Found")
	fmt.Println()
	fmt.Println("| Locale | Fields |")
	fmt.Println("|--------|--------|")
	for _, loc := range result.Localizations {
		fmt.Printf("| %s | %d |\n", loc.Locale, countNonEmptyFields(loc))
	}

	// App Info localizations
	if len(result.AppInfoLocalizations) > 0 {
		fmt.Println()
		fmt.Println("### App Info Localizations Found")
		fmt.Println()
		fmt.Println("| Locale | Name | Subtitle |")
		fmt.Println("|--------|------|----------|")
		for _, loc := range result.AppInfoLocalizations {
			name := "-"
			if loc.Name != "" {
				name = "✓"
			}
			subtitle := "-"
			if loc.Subtitle != "" {
				subtitle = "✓"
			}
			fmt.Printf("| %s | %s | %s |\n", loc.Locale, name, subtitle)
		}
	}

	if len(result.Uploaded) > 0 {
		fmt.Println()
		fmt.Println("### Uploaded")
		fmt.Println()
		for _, u := range result.Uploaded {
			fmt.Printf("- %s (%d fields)\n", u.Locale, u.Fields)
		}
	}

	if len(result.AppInfoUploaded) > 0 {
		fmt.Println()
		fmt.Println("### App Info Uploaded")
		fmt.Println()
		for _, u := range result.AppInfoUploaded {
			fmt.Printf("- %s (%d fields)\n", u.Locale, u.Fields)
		}
	}

	return nil
}

func printMigrateImportResultTable(result *MigrateImportResult) error {
	if result.DryRun {
		fmt.Println("DRY RUN - No changes made")
		fmt.Println()
	}
	fmt.Printf("Version ID: %s\n\n", result.VersionID)

	// Version localizations
	fmt.Println("Version Localizations:")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "LOCALE\tFIELDS\tSTATUS")
	for _, loc := range result.Localizations {
		status := "found"
		for _, u := range result.Uploaded {
			if u.Locale == loc.Locale {
				status = "uploaded"
				break
			}
		}
		fmt.Fprintf(w, "%s\t%d\t%s\n", loc.Locale, countNonEmptyFields(loc), status)
	}
	w.Flush()

	// App Info localizations
	if len(result.AppInfoLocalizations) > 0 {
		fmt.Println()
		fmt.Println("App Info Localizations:")
		w2 := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w2, "LOCALE\tNAME\tSUBTITLE\tSTATUS")
		for _, loc := range result.AppInfoLocalizations {
			status := "found"
			for _, u := range result.AppInfoUploaded {
				if u.Locale == loc.Locale {
					status = "uploaded"
					break
				}
			}
			name := "-"
			if loc.Name != "" {
				name = "yes"
			}
			subtitle := "-"
			if loc.Subtitle != "" {
				subtitle = "yes"
			}
			fmt.Fprintf(w2, "%s\t%s\t%s\t%s\n", loc.Locale, name, subtitle, status)
		}
		w2.Flush()
	}

	return nil
}

func printMigrateExportResultMarkdown(result *MigrateExportResult) error {
	fmt.Printf("**Version ID:** %s\n\n", result.VersionID)
	fmt.Printf("**Output Directory:** %s\n\n", result.OutputDir)
	fmt.Println("### Exported Locales")
	fmt.Println()
	for _, locale := range result.Locales {
		fmt.Printf("- %s\n", locale)
	}
	fmt.Printf("\n**Total Files:** %d\n", result.TotalFiles)
	return nil
}

func printMigrateExportResultTable(result *MigrateExportResult) error {
	fmt.Printf("Version ID: %s\n", result.VersionID)
	fmt.Printf("Output Dir: %s\n\n", result.OutputDir)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "LOCALE")
	for _, locale := range result.Locales {
		fmt.Fprintf(w, "%s\n", locale)
	}
	w.Flush()
	fmt.Printf("\nTotal Files: %d\n", result.TotalFiles)
	return nil
}
