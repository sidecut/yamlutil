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
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"github.com/sidecut/yamlutil/argsprocessor"
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
		err := argsprocessor.ProcessArgs(args, cmd.InOrStdin(), func(filename string, file io.Reader) {
			if filename != argsprocessor.STDIN {
				cmd.OutOrStdout().Write([]byte(fmt.Sprintf("%v:\n", filename)))
			}
			buf, err := ioutil.ReadAll(file)
			cobra.CheckErr(err)

			data := make(genericMap)
			err = yaml.Unmarshal(buf, &data)
			cobra.CheckErr(err)

			listKeys("", data)
		})
		cobra.CheckErr(err)
	},
}

// listKeys recursively lists all the keys in a map[string]interface{}
func listKeys(prefix string, data genericMap) {
	for key, value := range data {
		fmt.Printf("%v\n", fullKey(prefix, key))
		switch tValue := value.(type) {
		case string, bool:
			// do nothing
		case int:
			// Nothing -- don't drill down any further
		case []interface{}:
			listArray(fullKey(prefix, key), value.([]interface{}))
		case genericMap:
			listKeys(fullKey(prefix, key), tValue)
		case nil:
			// Nothing
		default:
			log.Fatalf("I don't know which type this value is: %v: %T", key, tValue)
		}
	}
}

func fullKey(prefix string, key any) string {
	if _, ok := key.(string); ok {
		return fmt.Sprintf("%v.%v", prefix, key)
	} else {
		return fmt.Sprintf("%v.\"%v\"", prefix, key)
	}
}

// listArray iterates through an array,
func listArray(key string, array []interface{}) {
	for i := range array {
		fmt.Printf("%v[%v]\n", key, i)
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
