// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"math/rand" //nolint
	"os"
	"time"

	"github.com/aunum/log"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/command/plugin"
)

var descriptor = cli.PluginDescriptor{
	Name:        "joesay",
	Description: "Stuff Joe says",
	Group:       cli.ExtraCmdGroup,
}

var stuffJoeSays = []string{
	"Perfect is the enemy of good",
	"Gotta crawl before you run",
	"Everybody's a somebody",
	"APIs are forever",
	"Kicking the tires",
	"Kubernetes is a platform platform",
	"You got to look at the big picture here",
	"That’s a Brendan special",
	"I didn’t invent Kubernetes so you kids could play Minecraft",
	"Off to the races",
	"Domain specific controller",
	"Goal seeking behavior",
	"Have you seen this, have you heard about this?",
	"We need to plumb that through",
	"We can lace that in",
	"Excited to see that get over the hump",
	"Pull down the yaml",
	"Spin up a kuard",
	"Want to see what it’s doing under the hood",
	"Well okay!",
	"Internet Explorer was great at its time",
}

func joeSay(say string) string {
	return `                                                                       
                                                                                
                 #,                                                             
                #                                                               
                .                                                               
                                                                                
               .                                                                
              .. .                                                              
             .                                                                  
             (,     ..                                                          
         ***(../  .... *,*%@@@@@@@@,/        . (/(&%##,                         
         /**,/* ..#@@.**.,,..    ,*,%@.      #@@%/    *%##.                     
          ,,//#..   %  .%@@,@@&(**/**/@,  @@  ..(@@%%,       ##    .            
          *,**(..   *         ,.,,,,.@     @  ... ...   .    .                        ` + fmt.Sprintf(`%q`, say) + `  
           / ,.       **         . (@       @               &                   
           .,,   .                                 ,*#(((*                      
             ,*,,....          ...                                              
             . **,,...      ,,,,*....,          .,,                             
              ,%**/,,,,....,*,,,,&@%@&,*@@@#%*    ,,.                           
              *,/#/,,,,,**@&*/,%#@@@@@@&@@@((,,..*((*                           
               *@##**/*,*@@,@@@@//(#**.,,&#@#*,/,#/(#*,,                        
                ,.%&&/.*(&@@@*,,**,#(     %*.....@@@@*./   .  .                 
                 .&/@%%#@@@@,*,,....            ,,@&%,/., ,                     
                  .**@(@%%( .#(((/(((//,#,*,    ... %#,&(                       
                    .(%,(,  (%%(,,*,, /.. .   . /  .&(/(.,                      
                      , ,(, .  *((**,,.,.,. ,       /# &                        
                     /.. &   . ...,/ **.*..         /     (                     
                     #.,,,,%*( *. ..%  @ .       .       .                      
                      & ,,,,***.@%&@&@@%.%#/(           .                       
                      .&/*,,,,*,,***,..  .            (                         
                       @.%*,,,,,,,. *,..            ., .                        
                        /*&*,,,.,,.....            *                            
                   .       */,...  .                 . @@@                      
                    , @@@@@ . *...             .  ,&@@@#@@**&.                 
                    %%@@  @@@@@*  .            .  &*, # ,  @@@@@@@             
             *    #&@ @@ ,    @/@* @@      .   ,@,  , &           .            
`
}

//nolint
func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}
	p.Cmd.RunE = func(cmd *cobra.Command, args []string) error {
		rand.Seed(time.Now().Unix())
		message := stuffJoeSays[rand.Intn(len(stuffJoeSays))]
		fmt.Println(joeSay(message))
		return nil
	}
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}
