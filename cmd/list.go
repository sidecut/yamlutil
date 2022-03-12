/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type genericMap map[interface{}]interface{}
type stringMap map[string]interface{}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all keys in the file",
	Run: func(cmd *cobra.Command, args []string) {
		switch len(args) {
		case 0:
			cobra.CheckErr(errors.New("Filename required"))

		default:
			for _, filename := range args {
				buf, err := os.ReadFile(filename)
				cobra.CheckErr(err)

				data := make(genericMap)
				yaml.Unmarshal(buf, &data)

				listKeys("", data)
			}
		}
	},
}

// listKeys recursively lists all the keys in a map[string]interface{}
func listKeys(prefix string, data genericMap) {
	for key, value := range data {
		fmt.Printf("%v\n", fullKey(prefix, key))
		switch t := value.(type) {
		case string:
			// do nothing
		case int:
			// Nothing -- don't drill down any further
		case []interface{}:
			listArray(prefix, key.(string), value.([]interface{}))
		case genericMap:
			listKeys(fullKey(prefix, key), t)
		default:
			log.Fatalf("I don't know which type this is: %v: %T", key, t)
		}
	}
}

func fullKey(prefix string, key any) string {
	return strings.Join([]string{prefix, key.(string)}, ".")
}

// listArray iterates through an array,
func listArray(prefix string, key string, array []interface{}) {
	for i, _ := range array {
		fmt.Printf("%v.%v[%v]\n", prefix, key, i)
	}
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
