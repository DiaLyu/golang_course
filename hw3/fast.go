package main 

import ( 
	"bufio" 
	"bytes" 
	"encoding/json" 
	"fmt" 
	"io" 
	"os" 
	"strings" 
) 

type User struct { 
	Name string `json:"name"` 
	Email string `json:"email"` 
	Browsers []string `json:"browsers"` 
} 

func FastSearch(out io.Writer) { 
	file, err := os.Open(filePath) 
	if err != nil { 
		panic(err) 
	} 
	defer file.Close() 
	
	scanner := bufio.NewScanner(file) 
	seenBrowsers := make(map[string]struct{}) 
	var foundUsers bytes.Buffer 
	userIndex := 0 
	
	for scanner.Scan() { 
		line := scanner.Bytes() 
		var user User 
		if err := json.Unmarshal(line, &user); err != nil { 
			panic(err) 
		} 
		
		isAndroid := false 
		isMSIE := false 
		
		for _, browser := range user.Browsers { 
			if strings.Contains(browser, "Android") { 
				isAndroid = true 
				seenBrowsers[browser] = struct{}{} 
			} 
			
			if strings.Contains(browser, "MSIE") { 
				isMSIE = true 
				seenBrowsers[browser] = struct{}{} 
			} 
		} 
		
		if isAndroid && isMSIE { 
			email := strings.Replace(user.Email, "@", " [at] ", 1) 
			fmt.Fprintf(&foundUsers, "[%d] %s <%s>\n", userIndex, user.Name, email) 
		} 
		
		userIndex++ 
	} 
	
	fmt.Fprintln(out, "found users:\n"+foundUsers.String()) 
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers)) 
}