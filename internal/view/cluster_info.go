package view

import (
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

var _ model.ClusterInfoListener = (*ClusterInfo)(nil)

// ClusterInfo represents a cluster info view.
type ClusterInfo struct {
	*tview.Table

	app    *App
	styles *config.Styles
}

// NewClusterInfo returns a new cluster info view.
func NewClusterInfo(app *App) *ClusterInfo {
	return &ClusterInfo{
		Table:  tview.NewTable(),
		app:    app,
		styles: app.Styles,
	}
}

// Init initializes the view.
func (c *ClusterInfo) Init() {
	c.SetBorderPadding(0, 0, 1, 0)
	c.app.Styles.AddListener(c)
	c.layout()
	c.StylesChanged(c.app.Styles)
}

// StylesChanged notifies skin changed.
func (c *ClusterInfo) StylesChanged(s *config.Styles) {
	c.styles = s
	c.SetBackgroundColor(s.BgColor())
	c.updateStyle()
}

func (c *ClusterInfo) layout() {
	for row, section := range []string{"Context", "Cluster", "User", "K9s Rev", "K8s Rev", "CPU", "MEM"} {
		if (section == "CPU" || section == "MEM") && !c.app.Conn().HasMetrics() {
			continue
		}
		c.SetCell(row, 0, c.sectionCell(section))
		c.SetCell(row, 1, c.infoCell(render.NAValue))
	}
}

func (c *ClusterInfo) sectionCell(t string) *tview.TableCell {
	cell := tview.NewTableCell(t + ":")
	cell.SetAlign(tview.AlignLeft)
	cell.SetBackgroundColor(tcell.ColorGreen)

	return cell
}

func (c *ClusterInfo) infoCell(t string) *tview.TableCell {
	cell := tview.NewTableCell(t)
	cell.SetExpansion(2)
	cell.SetTextColor(c.styles.K9s.Info.FgColor.Color())
	cell.SetBackgroundColor(c.app.Styles.BgColor())

	return cell
}

// ClusterInfoUpdated notifies the cluster meta was updated.
func (c *ClusterInfo) ClusterInfoUpdated(data model.ClusterMeta) {
	c.ClusterInfoChanged(data, data)
}

func (c *ClusterInfo) setCell(row int, s string) int {
	c.GetCell(row, 1).SetText(s)
	return row + 1
}

// ClusterInfoChanged notifies the cluster meta was changed.
func (c *ClusterInfo) ClusterInfoChanged(prev, curr model.ClusterMeta) {
	c.app.QueueUpdateDraw(func() {
		c.Clear()
		c.layout()
		row := c.setCell(0, curr.Context)
		row = c.setCell(row, curr.Cluster)
		row = c.setCell(row, curr.User)
		row = c.setCell(row, curr.K9sVer)
		row = c.setCell(row, curr.K8sVer)
		if c.app.Conn().HasMetrics() {
			row = c.setCell(row, ui.AsPercDelta(prev.Cpu, curr.Cpu))
			_ = c.setCell(row, ui.AsPercDelta(prev.Mem, curr.Mem))
			var set bool
			if c.app.Config.K9s.Thresholds.ExceedsCPUPerc(curr.Cpu) {
				c.app.Status(model.FlashErr, "CPU on fire!")
				set = true
			}
			if c.app.Config.K9s.Thresholds.ExceedsMemoryPerc(curr.Mem) {
				c.app.Status(model.FlashErr, "Memory on fire!")
				set = true
			}
			if !set {
				c.app.ClearStatus(true)
			}
		}
		c.updateStyle()
	})
}

func (c *ClusterInfo) updateStyle() {
	for row := 0; row < c.GetRowCount(); row++ {
		c.GetCell(row, 0).SetTextColor(c.styles.K9s.Info.FgColor.Color())
		c.GetCell(row, 0).SetBackgroundColor(c.styles.BgColor())
		var s tcell.Style
		c.GetCell(row, 1).SetStyle(s.Bold(true).Foreground(c.styles.K9s.Info.SectionColor.Color()))
	}
}
