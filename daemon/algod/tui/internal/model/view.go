package model

import (
	"github.com/algorand/go-algorand/daemon/algod/tui/internal/constants"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/indent"
)

// TODO: this function could implement a type and be passed to the tab view.
func (m Model) tabView() string {
	switch activeComponent(m.Tabs.GetActiveIndex()) {
	case explorerTab:
		return m.BlockExplorer.View()
	case configTab:
		return m.Configs.View()
	case helpTab:
		return m.About.View()
	}

	return "unknown tab"
}

func art() string {
	// TODO: This could take a width and indent/border to line up with the bottom
	/*
					art := `

				         ////\\
				        ////\\\\
				       ////  \\////
				      ////    ////
				     ////    ////\\
				    ////    ////\\\\
				   ////    ////  \\\\
				  ////    ////    \\\\
				 ////    ////      \\\\
				////    ////        \\\\
				`

			art := `
		               .%@@@@@@@%
		              :@@@@@@@@@@=
		             -@@@@@=:@@@@@****=
		            =@@@@@-  *@@@@@@@#
		           +@@@@@:   .@@@@@@*
		          *@@@@%.    -@@@@@#
		        .%@@@@%.    =@@@@@@@.
		       .@@@@@#     +@@@@@@@@*
		      :@@@@@+     #@@@@%@@@@@.
		     -@@@@@=    .%@@@@%.=@@@@#
		    =@@@@@-    .%@@@@*   @@@@@:
		   *@@@@%.    :@@@@@+    =@@@@%
		  #@@@@%.    -@@@@@=      %@@@@-
		.%@@@@#     +@@@@@-       -@@@@%`
	*/
	art := `
                ███████ 
              ██████████▒
            ▒█████▒ █████▓▓▓▓▒
           ▒█████▒  ▓███████▓
          ▓█████     ██████▓
         ▓█████     ▒█████▓
        ██████     ▒███████
       █████▓     ▓████████▓
     ▒█████▓     ▓██████████
    ▒█████▒     ██████ ▒████▓
   ▒█████▒     █████▓   █████▒
  ▓█████     ▒█████▓    ▒█████▒`
	return indent.String(art, 3)
}

func (m Model) View() string {
	// Compose the different views by joining them together in the right orientation.
	return lipgloss.JoinVertical(0,
		lipgloss.JoinHorizontal(0,
			m.Status.View(),
			art()),
		//m.Accounts.View()),
		m.Tabs.View(),
		m.tabView(),
		m.Help.View(constants.Keys),
		m.Footer.View())
}
