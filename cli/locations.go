package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/dangrier/ocm-client/ocm"
	"github.com/spf13/cobra"
)

var locationsTypeFlag string

var locationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "List all known locations",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := clientWithTimeout()
		defer cancel()

		locs, err := s.client.GetLocations(ctx)
		if err != nil {
			return err
		}

		list := make([]*ocm.Location, 0, len(locs))
		for _, l := range locs {
			if locationsTypeFlag != "" {
				lt, err := ocm.ParseLocationType(locationsTypeFlag)
				if err != nil {
					return err
				}
				if l.Type != lt {
					continue
				}
			}
			list = append(list, l)
		}

		return writeLocations(s.output, list)
	},
}

var locationCmd = &cobra.Command{
	Use:   "location",
	Short: "Show a single location by code",
	RunE: func(cmd *cobra.Command, args []string) error {
		code, _ := cmd.Flags().GetInt("code")
		ctx, cancel := clientWithTimeout()
		defer cancel()

		l, err := s.client.GetLocation(ctx, code)
		if err != nil {
			return err
		}
		return writeLocations(s.output, []*ocm.Location{l})
	},
}

func init() {
	locationsCmd.Flags().StringVarP(&locationsTypeFlag, "type", "t", "", "Filter by location type: suburb, postcode, lga, nhw, region, district, patrol, division")
	locationCmd.Flags().Int("code", 0, "Location code")
	_ = locationCmd.MarkFlagRequired("code")
}

func writeLocations(format string, locs []*ocm.Location) error {
	switch strings.ToLower(format) {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(locs)
	default:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, err := fmt.Fprintln(w, "CODE\tTYPE\tNAME")
		if err != nil {
			return err
		}
		for _, l := range locs {
			_, err := fmt.Fprintf(w, "%d\t%s\t%s\n", l.Code, l.Type, l.Name)
			if err != nil {
				return err
			}
		}
		return w.Flush()
	}
}
