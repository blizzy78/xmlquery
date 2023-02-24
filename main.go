package main

import (
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/mattn/go-isatty"
	"github.com/ttacon/chalk"
)

type options struct {
	expression   string
	contentsOnly bool
	color        bool
	noColor      bool
	noChildren   bool
	printHelp    bool
}

type commandLineError string

type stringWriter interface {
	io.Writer
	io.StringWriter
}

func main() {
	if err := run(); err != nil {
		var cmdLineError commandLineError
		if errors.As(err, &cmdLineError) {
			fmt.Fprintln(os.Stderr, cmdLineError.Error())
			fmt.Fprintln(os.Stderr)

			flag.Usage()

			os.Exit(1)
		}

		panic(err)
	}
}

func run() error {
	opts, err := parseOptions()
	if err != nil {
		return err
	}

	doc, err := xmlquery.Parse(os.Stdin)
	if err != nil {
		return fmt.Errorf("parse XML: %w", err)
	}

	isTerminal := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	useColor := isTerminal || opts.color

	if opts.noColor {
		useColor = false
	}

	nodes, err := xmlquery.QueryAll(doc, opts.expression)
	if err != nil {
		return fmt.Errorf("execute query: %w", err)
	}

	for _, node := range nodes {
		if err := outputXML(os.Stdout, node, !opts.contentsOnly, useColor, !opts.noChildren); err != nil {
			return err
		}

		if _, err := os.Stdout.WriteString("\n"); err != nil {
			return fmt.Errorf("write XML: %w", err)
		}
	}

	return nil
}

func parseOptions() (options, error) {
	opts := options{}

	flag.BoolVar(&opts.printHelp, "help", opts.printHelp, "print this help")
	flag.StringVar(&opts.expression, "expr", opts.expression, "XPath expression to select nodes from the input")
	flag.BoolVar(&opts.contentsOnly, "contents-only", opts.contentsOnly, "print only the contents of selected nodes")
	flag.BoolVar(&opts.color, "color", opts.color, "use colored output")
	flag.BoolVar(&opts.noColor, "no-color", opts.noColor, "don't use colored output")
	flag.BoolVar(&opts.noChildren, "no-children", opts.noChildren, "don't output child nodes of selected nodes")

	flag.Parse()

	if opts.printHelp {
		flag.Usage()

		os.Exit(0)

		return options{}, nil
	}

	if opts.expression == "" {
		return options{}, commandLineError("no expression specified")
	}

	if opts.color && opts.noColor {
		return options{}, commandLineError("cannot use -color and -no-color together")
	}

	return opts, nil
}

func outputXML(writer stringWriter, node *xmlquery.Node, self bool, color bool, recursive bool) error {
	if self {
		if err := outputXMLToBuffer(writer, node, color, recursive); err != nil {
			return err
		}

		return nil
	}

	for n := node.FirstChild; n != nil; n = n.NextSibling {
		if err := outputXMLToBuffer(writer, n, color, recursive); err != nil {
			return err
		}
	}

	return nil
}

func outputXMLToBuffer(writer stringWriter, node *xmlquery.Node, color bool, recursive bool) error { //nolint:gocognit,cyclop // it's a bit complicated
	fullNodeName := node.Data
	if node.Prefix != "" {
		fullNodeName = node.Prefix + ":" + node.Data
	}

	if node.Type == xmlquery.TextNode || node.Type == xmlquery.CommentNode {
		if err := xml.EscapeText(writer, []byte(strings.TrimSpace(node.Data))); err != nil {
			return fmt.Errorf("write XML: %w", err)
		}

		return nil
	}

	if node.Type == xmlquery.DeclarationNode { //nolint:nestif // it's a bit complicated
		if color {
			if _, err := writer.WriteString(chalk.Green.String()); err != nil {
				return fmt.Errorf("write color code: %w", err)
			}
		}

		if _, err := writer.WriteString("<?" + node.Data); err != nil {
			return fmt.Errorf("write XML: %w", err)
		}
	} else {
		if color {
			if _, err := writer.WriteString(chalk.Magenta.String()); err != nil {
				return fmt.Errorf("write color code: %w", err)
			}
		}

		if _, err := writer.WriteString("<" + fullNodeName); err != nil {
			return fmt.Errorf("write XML: %w", err)
		}
	}

	if color {
		if _, err := writer.WriteString(chalk.Reset.String()); err != nil {
			return fmt.Errorf("write color code: %w", err)
		}
	}

	for _, attr := range node.Attr {
		if color {
			if _, err := writer.WriteString(chalk.Yellow.String()); err != nil {
				return fmt.Errorf("write color code: %w", err)
			}
		}

		attrName := attr.Name.Local
		if attr.Name.Space != "" {
			attrName = attr.Name.Space + ":" + attr.Name.Local
		}

		if _, err := writer.WriteString(" " + attrName + `="`); err != nil {
			return fmt.Errorf("write XML: %w", err)
		}

		if color {
			if _, err := writer.WriteString(chalk.Reset.String() + chalk.Cyan.String()); err != nil {
				return fmt.Errorf("write color code: %w", err)
			}
		}

		if _, err := writer.WriteString(attr.Value); err != nil {
			return fmt.Errorf("write XML: %w", err)
		}

		if color {
			if _, err := writer.WriteString(chalk.Reset.String() + chalk.Yellow.String()); err != nil {
				return fmt.Errorf("write color code: %w", err)
			}
		}

		if _, err := writer.WriteString(`"`); err != nil {
			return fmt.Errorf("write XML: %w", err)
		}
	}

	if color {
		if _, err := writer.WriteString(chalk.Reset.String()); err != nil {
			return fmt.Errorf("write color code: %w", err)
		}
	}

	if node.Type == xmlquery.DeclarationNode { //nolint:nestif // it's a bit complicated
		if color {
			if _, err := writer.WriteString(chalk.Green.String()); err != nil {
				return fmt.Errorf("write color code: %w", err)
			}
		}

		if _, err := writer.WriteString("?>"); err != nil {
			return fmt.Errorf("write XML: %w", err)
		}
	} else {
		if color {
			if _, err := writer.WriteString(chalk.Magenta.String()); err != nil {
				return fmt.Errorf("write color code: %w", err)
			}
		}

		if _, err := writer.WriteString(">"); err != nil {
			return fmt.Errorf("write XML: %w", err)
		}
	}

	if color {
		if _, err := writer.WriteString(chalk.Reset.String()); err != nil {
			return fmt.Errorf("write color code: %w", err)
		}
	}

	if recursive {
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if err := outputXMLToBuffer(writer, child, color, recursive); err != nil {
				return err
			}
		}
	}

	if node.Type != xmlquery.DeclarationNode { //nolint:nestif // it's a bit complicated
		if color {
			if _, err := writer.WriteString(chalk.Magenta.String()); err != nil {
				return fmt.Errorf("write color code: %w", err)
			}
		}

		if _, err := writer.WriteString("</" + fullNodeName + ">"); err != nil {
			return fmt.Errorf("write XML: %w", err)
		}

		if color {
			if _, err := writer.WriteString(chalk.ResetColor.String()); err != nil {
				return fmt.Errorf("write color code: %w", err)
			}
		}
	}

	return nil
}

func (e commandLineError) Error() string {
	return string(e)
}
