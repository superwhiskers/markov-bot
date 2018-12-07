/*

hemlock-parser - program that parses hemlock message database files
Copyright (C) 2018 superwhiskers <whiskerdev@protonmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

*/

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"encoding/json"
	"strconv"
	//"strings"
)

var (
	// command line flag variables
	dbFilePath     *string
	outputFilePath *string

	// file data variables
	dbFileData []hemlockContent

	err error
)

func init() {

	dbFilePath = flag.String("database", "hemlock.json", "the path to the hemlock database file")
	outputFilePath = flag.String("output", "hemlock.json.out", "the path to the file to output the edited database to")

}

func main() {

	flag.Parse()

	fmt.Printf("-- loading %s\n", *dbFilePath)

	dbFileByte, err := ioutil.ReadFile(*dbFilePath)
	if err != nil {

		fmt.Printf("-- error while reading %s. error: %v\n", *dbFilePath, err)
		os.Exit(1)

	}

	err = json.Unmarshal(dbFileByte, &dbFileData)
	if err != nil {

		fmt.Printf("-- invalid json in %s. error: %v\n", *dbFilePath, err)
		os.Exit(1)

	}

	fmt.Printf("-- loaded %d entries from %s\n", len(dbFileData), *dbFilePath)

	for {

		fmt.Println()
		fmt.Println("hemlock-parser by superwhiskers")
		fmt.Println("version zero")
		fmt.Println("----")
		fmt.Println()
		fmt.Println("  1. rate unrated messages")
		fmt.Println("  2. find data in the db")
		fmt.Println("  3. output changes")
		fmt.Println("  4. merge databases")
		fmt.Println("  5. exit")
		fmt.Println()

		switch question("select an option", []string{"1", "2", "3", "4", "5"}) {

		case "1":

			var rating string
			for i, content := range dbFileData {

				if content.Rating != -1 {

					continue

				}

				fmt.Printf("\n")
				fmt.Printf("-- rating content %d/%d\n", i+1, len(dbFileData))
				fmt.Printf("content: `%s`\n", content.Content)
				rating = question("how would you rate this?", []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "x", "i"})

				if rating == "x" {

					break

				}

				switch rating {

				case "1", "2", "3", "4", "5", "6", "7", "8", "9", "10":
					dbFileData[i].Rating, _ = strconv.Atoi(rating)

				case "i":
					continue

				}

			}

		case "2":
			fmt.Println("-- not implemented")

		case "3":
			outputFileData, err := json.Marshal(dbFileData)
			if err != nil {

				fmt.Printf("-- error while stringifying json. error: %v\n", err)
				os.Exit(1)

			}

			err = ioutil.WriteFile(*outputFilePath, outputFileData, 0644)
			if err != nil {

				fmt.Printf("-- error while outputting data to %s. error: %v\n", *outputFilePath, err)
				os.Exit(1)

			}
			fmt.Printf("-- finished writing modified data to %s\n", *outputFilePath)

		case "4":
			//files := strings.Split(question("which files do you want to merge? (separate each name with a comma)", []string{}), ",")

			/*var (
				fileDataByte []byte
				fileData []hemlockData
				err error
			)

			for _, file := range files {

				fileData, err = ioutil.ReadFile(file)
				if err != nil {

					fmt.Printf("-- unable to read file. error: %v (ignoring error)", err)
					continue

				}

				fmt.Printf("-- merging %s with the currently loaded file", file)

				for i, selectedFileContent := range fileData {

					if dbFileData[i].Content == selectedFileContent.Content {*/


		case "5":
			os.Exit(0)

		}

	}

}
