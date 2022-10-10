/**
Tools package. It's contain some useful tools, just like vote and so on.
**/
package tools

import (
	"fmt"
	"os"
	"io/ioutil"
	"strings"
	"strconv"
	"github.com/urfave/cli"
	"github.com/outbrain/golib/log"
	"github.com/csunny/dpos"

)

// NodeVote 
var NodeVote = cli.Command{
	Name: "vote",
	Usage: "vote for node",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "name",
			Value: "",
			Usage: "",
		},
		cli.IntFlag{
			Name: "v",
			Value: 0,
			Usage: "",
		},
	}, 
	Action: func(context *cli.Context) error{
		if err := Vote(context); err != nil{
			return err
		} 
		return nil
	},
}
// Vote for node. The votes of node is origin vote plus new vote.
// votes = originVote + vote 
func Vote(context *cli.Context) error {
	name := context.String("name")
	vote := context.Int("v")

	if name == "" {
		log.Errorf("")
	}

	if vote < 1 {
		log.Errorf("")
	}

	f, err := ioutil.ReadFile(dpos.FileName)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	res := strings.Split(string(f), "\n")

	voteMap := make(map[string]string)
	for _, node := range res {
		nodeSplit := strings.Split(node, ":")
		if len(nodeSplit) > 1 {
			voteMap[nodeSplit[0]] = fmt.Sprintf("%s", nodeSplit[1])
		}
	}

	originVote, err := strconv.Atoi(voteMap[name])
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	votes := originVote + vote
	voteMap[name] = fmt.Sprintf("%d", votes)

	log.Infof("%s%d", name, vote)
	str := ""
	for k, v := range voteMap {
		str += k + ":" + v + "\n"
	}

	file, err := os.OpenFile(dpos.FileName, os.O_RDWR, 0666)
	if err != nil{
		return err
	}

	file.WriteString(str)
	defer file.Close()

	return nil
}