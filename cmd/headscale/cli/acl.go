package cli

import (
	"fmt"

	v1 "github.com/juanfont/headscale/gen/go/headscale/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/status"
)

func init() {
	//	rootCmd.AddCommand(userCmd)
	//	userCmd.AddCommand(listUsersCmd)
	rootCmd.AddCommand(aclCmd)
	aclCmd.AddCommand(aclPingPongCmd)
}

var aclCmd = &cobra.Command{
	Use:   "acl",
	Short: "Manage the acl of Headscale",
}

var aclPingPongCmd = &cobra.Command{
	Use:   "ping",
	Short: "Do PingPong test for ACL gRPC test",
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.Flags().GetString("output")

		ctx, client, conn, cancel := getHeadscaleCLIClient()
		defer cancel()
		defer conn.Close()

		request := &v1.ACLPingPongRequest{PingMsg: "ping"}

		response, err := client.ACLPingPong(ctx, request)
		if err != nil {
			ErrorOutput(
				err,
				fmt.Sprintf("Cannot do PingPong: %s", status.Convert(err).Message()),
				output,
			)

			return
		}

		SuccessOutput(response.PongMsg, response.PongMsg, "")

	},
}
