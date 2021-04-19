package cli

import (
	"archive/zip"
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check which plugins have updates.",
	Run: func(cmd *cobra.Command, args []string) {
		m := getBundledPlugins()

		pluginsToUpdate := make(map[string]string)

		var wg sync.WaitGroup

		wg.Add(len(m))

		for k, v := range m {
			go func(pluginName string, currentVersion string) {
				defer wg.Done()

				fmt.Println(currentVersion)
				plugin, err := getPlugin(pluginName)
				if err != nil {
					panic(err)
				}

				latestVersion := plugin.Version

				fmt.Println("Latest Version: " + latestVersion)

				fp := filepath.Join("plugins", pluginName+".jar")

				reader, err := zip.OpenReader(fp)

				if err != nil {
					panic(err)
				}

				for _, file := range reader.File {
					if strings.HasSuffix(file.Name, "plugin.yml") {
						yml := &PluginYML{}
						rc, err := file.Open()
						if err != nil {
							panic(err)
						}
						buf := bytes.Buffer{}
						buf.ReadFrom(rc)
						yaml.Unmarshal(buf.Bytes(), yml)

						if yml.Version != latestVersion {
							pluginsToUpdate[pluginName] = latestVersion
						}
					}
				}
			}(k, v)
		}
		wg.Wait()
		if len(pluginsToUpdate) != 0 {
			fmt.Println("Plugins To Update:")
			for k, v := range pluginsToUpdate {
				fmt.Println(k, " -> ", v)
			}
		} else {
			fmt.Println("All plugins are up to date :)")
		}

	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
