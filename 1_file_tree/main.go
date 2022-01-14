package main

import (
	"errors"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

func isFileIgnored(fi os.FileInfo, printFiles bool) bool {
	if fi.Name() == ".DS_Store" {
		return true
	}
	if !fi.IsDir() && !printFiles {
		return true
	}
	return false
}

func formatFileSize(size int64) string {
	var sb strings.Builder
	sb.WriteString(" (")
	if size == 0 {
		sb.WriteString("empty")
		sb.WriteString(")")
	} else {
		sb.WriteString(strconv.FormatInt(size, 10))
		sb.WriteString("b)")
	}
	return sb.String()
}

func getNextLevelPrefix(prefix string, isLast bool) string {
	if !isLast {
		return prefix + "│\t"
	} else {
		return prefix + "\t"
	}
}

func getNextLevelPath(fi os.FileInfo, path string) string {
	return path + string(os.PathSeparator) + fi.Name()
}

func formatCurrLevelPrefix(prefix string, isLast bool) string {
	if !isLast {
		return prefix + "├───"
	} else {
		return prefix + "└───"
	}
}

func formatFileStats(fi os.FileInfo, prefix string, isLast bool) string {
	var sb strings.Builder
	sb.WriteString(formatCurrLevelPrefix(prefix, isLast))
	sb.WriteString(fi.Name())
	if !fi.IsDir() {
		sb.WriteString(formatFileSize(fi.Size()))
	}
	sb.WriteString("\n")
	return sb.String()
}

func dirTreeWalk(w io.Writer, path string, dir_prefix string, printFiles bool) error {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return errors.New("file must be dir")
	}
	fis, err := f.Readdir(0)
	if err != nil {
		return err
	}

	filtered_fis := []os.FileInfo{}
	for _, fi := range fis {
		if !isFileIgnored(fi, printFiles) {
			filtered_fis = append(filtered_fis, fi)
		}
	}
	fis = filtered_fis
	sort.Slice(fis, func(i, j int) bool {
		return fis[i].Name() < fis[j].Name()
	})

	for i, fi := range filtered_fis {
		var isLast = (i+1 == len(filtered_fis))
		w.Write([]byte(formatFileStats(fi, dir_prefix, isLast)))
		if fi.IsDir() {
			dirTreeWalk(w, getNextLevelPath(fi, path), getNextLevelPrefix(dir_prefix, isLast), printFiles)
		}
	}
	return nil
}

func dirTree(w io.Writer, path string, printFiles bool) error {
	if err := dirTreeWalk(w, path, "", printFiles); err != nil {
		return err
	}
	return nil
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
