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
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type genericMap map[interface{}]interface{}
type stringMap map[string]interface{}
type interfaceArray []interface{}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all keys in the file",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.PrintErrln("Filename required")
			return
		}

		filename := args[0]
		buf, err := os.ReadFile(filename)
		if err != nil {
			cmd.PrintErrf("%v\n", err)
		}

		// yaml := string(buf)
		// fmt.Printf("%v\n", yaml)
		data := make(stringMap)
		yaml.Unmarshal(buf, &data)

		listKeys(".", data)
	},
}

// listKeys recursively lists all the keys in a map[string]interface{}
func listKeys(prefix string, data stringMap) {
	for key, value := range data {
		fmt.Printf("%v\n", strings.Join([]string{prefix, key}, "/"))
		switch t := value.(type) {
		case string:
			// do nothing
		case stringMap:
			// log.Println("Recursing")
			listKeys(prefix+"/"+key, t)
		case interfaceArray:
			// This is an array of things
			listArray(prefix, value.(interfaceArray))
		default:
			log.Fatalf("I don't know which type this is: %T", t)
			// panic("I don't know which type this is")
		}
	}
}

// listArray iterates through an array,
func listArray(prefix string, array interfaceArray) {

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
