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
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var filename string // means sort in-place
var infilename string
var outfilename string
var useStdOut bool

const eitherButNotBothErrorMessage = "--in or -f/--file must be specified but not both"

// sortCmd represents the sort command
var sortCmd = &cobra.Command{
	Use:   "sort",
	Short: "Sort YAML keys",
	Long:  eitherButNotBothErrorMessage,
	Run: func(cmd *cobra.Command, args []string) {
		err := validateParameters()
		cobra.CheckErr(err)

		var yamlFile []byte
		yamlFile, err = ioutil.ReadFile(infilename)
		if err != nil {
			log.Fatalln(err)
		}
		var yamlContents map[string]interface{}
		err = yaml.Unmarshal(yamlFile, &yamlContents)
		if err != nil {
			log.Fatalln(err)
		}
		// fmt.Printf("%#v", yamlContents)
		outBuffer, err := yaml.Marshal(yamlContents)
		if err != nil {
			log.Fatalln(err)
		}

		if useStdOut {
			if _, err := cmd.OutOrStdout().Write(outBuffer); err != nil {
				log.Fatalln(err)
			}
		} else {
			if err := ioutil.WriteFile(outfilename, outBuffer, 0644); err != nil {
				log.Fatalln(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(sortCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sortCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	sortCmd.Flags().StringVar(&infilename, "in", "", "Input filename.  Must not be used with -f/--file.")
	sortCmd.Flags().StringVar(&outfilename, "out", "", `Output filename.  Must be used with --in.  Must not be used with -f/--file.
If omitted, writes to stdout.`)
	sortCmd.Flags().StringVarP(&filename, "file", "f", "", "Input and output filename; sorts in place.  Cannot be used with --in or --out.")
}

func validateParameters() (err error) {
	// Truth table
	if filename == "" && infilename == "" {
		// Must use either --file or --in
		err = errors.New(eitherButNotBothErrorMessage)
	} else if filename == "" && infilename != "" {
		// infilename has been specified, and use stdout if no outfilename has been specified
		useStdOut = outfilename == ""
	} else if filename != "" && infilename == "" && outfilename == "" {
		infilename = filename
		outfilename = filename
	} else if filename != "" && infilename == "" && outfilename != "" {
		err = errors.New("-f/--file must be used alone")
	} else if filename != "" && infilename != "" {
		err = errors.New(eitherButNotBothErrorMessage)
	}

	return
}
