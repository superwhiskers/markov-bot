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
	"bufio"
	"fmt"
	"strings"
	"os"
)

// function that asks a question until it gets a valid answer
func question(prompt string, valid []string) string {

	var inp string

	for {

		fmt.Printf("%s\n", prompt)
		if len(valid) != 0 {

			fmt.Printf("(%s): ", strings.Join(valid, ", "))

		} else {

			fmt.Printf(": ")

		}

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		inp = scanner.Text()

		if len(valid) == 0 {

			return inp

		}

		for _, ele := range valid {

			if ele == inp {

				return inp

			}

		}

		fmt.Printf("\"%s\" is not a valid answer\n", inp)

	}

}
