## web-scraper
#### The tool written in Go to download files in sequential and concurrent mode
#### Download all html, image, css and js files with keeping the file structure of the website
1. Clone the repository
2. Run the script using the command `go run main.go` or build the binary using `go build main.go` and run the binary `./web-scraper`
3. To be able to see downloaded website you can visit following page in the browser `file://{absolut_path_to_the_script}/books.toscrape.com/index.html`
4. Get absolute path to the script by running `pwd` in the terminal where you run the script
5. The script is using following libraries:
   - `goquery` library to parse HTML
   - `progressbar` library to show progress bar
   - `net/http` standard library to make HTTP requests
   - `testing` library to write tests
6. The logic of the script is following:
   - get the website URL from the const
   - parse the website content to get all links to the files, store them in the slice
   - create the directory with the name of the website
   - download all files from the website using the slice with links
   - download files in sequential mode and concurrent mode
   - display the results (progress bar, time and amount of downloaded files) for sequential and concurrent mode
7. The script implements following ways to download files from the website:
   - sequentially downloading files from the website - traverse the website and download files one by one
   - in concurrent mode by `sync.WaitGroup` to wait for all goroutines to finish and using `channels` to communicate between goroutines
8. As result of the script execution you will get:
   - `books.toscrape.com` directory with all downloaded files in current directory
   - the message how many files going to be downloaded
   - progress bar showing the progress of downloading files for sequential and concurrent mode
   - the message with time of downloading files for sequential and concurrent mode
   - the message with a number of downloaded files
9. `main_test.go` contains tests for the script - to run tests use `go test` command
