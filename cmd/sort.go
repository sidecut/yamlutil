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
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var replace bool
var automaticName bool
var useStdIn bool

const eitherButNotBothErrorMessage = "--in or --file must be specified but not both"

// Usage:
// filenames are files to sort
// -r --replace -- do an in-place sort
// -a --auto -- automatic *.out.yaml filename
// If no filenames, use stdin

// sortCmd represents the sort command
var sortCmd = &cobra.Command{
	Use:   "sort",
	Short: "Sort YAML keys",
	Long: `Sorts YAML keys

Non-option arguments are names of files to sort.
If no filenames, use stdin.

-r --replace -- do an in-place sort
-a --auto -- automatic *.out.yaml filename`,
	Run: func(cmd *cobra.Command, args []string) {
		if automaticName && replace {
			err := errors.New("--auto and --replace cannot both be used")
			cobra.CheckErr(err)
		}

		// _, err := cmd.OutOrStdout().Write([]byte(fmt.Sprintf("args: %#v", args)))
		// cobra.CheckErr(err)

		if len(args) == 0 {
			if automaticName || replace {
				err := errors.New("Can't use --auto or --replace with stdin")
				cobra.CheckErr(err)
			}

			doSort(cmd, cmd.InOrStdin(), cmd.OutOrStdout())
		} else {
			for _, filename := range args {
				func() {
					input, output, err := getInputAndOutputForFilename(cmd, filename)
					defer input.Close()
					defer output.Close()
					if err != nil {
						cmd.PrintErrln(err)
						return
					}
					err = doSort(cmd, input, output)
					if err != nil {
						cmd.PrintErrln(err)
						return
					}
				}()
			}
		}

	},
}

func getInputAndOutputForFilename(cmd *cobra.Command, filename string) (input *os.File, output *os.File, err error) {
	input, err = os.Open(filename)
	if err != nil {
		return
	}

	if automaticName {
		outputFilename := ""
		outputFilename, err = getOutputFilename(filename)
		if err != nil {
			return
		}
		output, err = os.Create(outputFilename)
		if err != nil {
			return
		}
	} else {
		output = cmd.OutOrStdout().(*os.File)
	}

	return
}

func getOutputFilename(filename string) (string, error) {
	const out_yaml = ".out.yaml"
	parts := strings.Split(filename, ".")
	switch len(parts) {
	case 0:
		return "", errors.New("Empty filename")
	case 1:
		return filename + out_yaml, nil
	default:
		extension := parts[len(parts)-1]
		if strings.ToLower(extension) == "yaml" {
			return strings.Join(parts[:len(parts)-1], ".") + out_yaml, nil
		} else {
			return filename + out_yaml, nil
		}
	}
}

func doSort(cmd *cobra.Command, input io.Reader, output io.Writer) (err error) {
	//var yamlBytes []byte
	yamlBytes, err := ioutil.ReadAll(input) //.ReadFile(infilename)
	if err != nil {
		return
	}

	var yamlMap map[string]interface{}
	err = yaml.Unmarshal(yamlBytes, &yamlMap)
	if err != nil {
		return
	}
	// fmt.Printf("%#v", yamlContents)
	outBuffer, err := yaml.Marshal(yamlMap)
	if err != nil {
		return
	}

	if _, err = cmd.OutOrStdout().Write(outBuffer); err != nil {
		return
	}

	return
}

func init() {
	rootCmd.AddCommand(sortCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sortCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	sortCmd.Flags().BoolVarP(&replace, "replace", "r", false, "Do an in-place sort, replacing the file(s).")
	sortCmd.Flags().BoolVarP(&automaticName, "auto", "a", false, "Automatic *.out.yaml filename.")
}
