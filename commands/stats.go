package commands

import (
	cmds "gx/ipfs/QmQtQrtNioesAWtrx8csBvfY37gTe94d6wQ3VikZUjxD39/go-ipfs-cmds"
	cmdkit "gx/ipfs/Qmde5VP1qUkyQXKCfmEUA7bP64V2HAptbJ7phuPp7jXWwg/go-ipfs-cmdkit"

	api "github.com/filecoin-project/go-filecoin/api"
)

var statsCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "View various filecoin node statistics",
	},
	Subcommands: map[string]*cmds.Command{
		"bw": statsBwCmd,
	},
}

var statsBwCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "view bandwidth usage metrics",
	},
	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("protocol", false, false, "protocol name to list metrics for"),
	},
	Run: func(req *cmds.Request, re cmds.ResponseEmitter, env cmds.Environment) error {
		var proto string
		if len(req.Arguments) > 0 {
			proto = req.Arguments[0]
		}

		bwstats := GetAPI(env).Stats().Bandwidth(req.Context, proto)

		return re.Emit(bwstats)
	},
	Type: api.BWStats{},
}
