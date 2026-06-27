package admin

import (
	"strings"

	"github.com/cybersaksham/gogo/auth"
	gogohttp "github.com/cybersaksham/gogo/http"
)

// IndexContext is the render-ready admin index data.
type IndexContext struct {
	Site *Site
	Apps []IndexApp
}

// IndexApp groups visible models by app label.
type IndexApp struct {
	AppLabel string
	Models   []IndexModel
}

// IndexModel describes one visible model entry.
type IndexModel struct {
	AppLabel  string
	Name      string
	AddURL    string
	ChangeURL string
}

// BuildIndex builds the permission-filtered admin index.
func BuildIndex(site *Site, router *gogohttp.Router, user auth.User) (IndexContext, error) {
	return buildIndex(site, router, user, "")
}

// BuildAppList builds a permission-filtered app list for one app label.
func BuildAppList(site *Site, router *gogohttp.Router, user auth.User, appLabel string) (IndexContext, error) {
	return buildIndex(site, router, user, strings.ToLower(appLabel))
}

func buildIndex(site *Site, router *gogohttp.Router, user auth.User, onlyApp string) (IndexContext, error) {
	if site == nil {
		site = DefaultSite()
	}
	if router == nil {
		var err error
		router, err = site.URLs()
		if err != nil {
			return IndexContext{}, err
		}
	}
	context := IndexContext{Site: site}
	appPositions := map[string]int{}
	for _, label := range site.ModelRegistry.RegisteredModels() {
		admin, ok := site.ModelRegistry.GetAdmin(label)
		if !ok {
			continue
		}
		appLabel := strings.ToLower(admin.Model.AppLabel)
		modelName := strings.ToLower(admin.Model.ModelName)
		if onlyApp != "" && appLabel != onlyApp {
			continue
		}
		if !canViewOrChange(user, appLabel, modelName) {
			continue
		}
		changeURL, err := router.Reverse(routeName(appLabel, modelName, "changelist"), nil)
		if err != nil {
			return IndexContext{}, err
		}
		addURL := ""
		if auth.HasPerm(user, appLabel+".add_"+modelName) {
			addURL, err = router.Reverse(routeName(appLabel, modelName, "add"), nil)
			if err != nil {
				return IndexContext{}, err
			}
		}
		position, ok := appPositions[appLabel]
		if !ok {
			context.Apps = append(context.Apps, IndexApp{AppLabel: appLabel})
			position = len(context.Apps) - 1
			appPositions[appLabel] = position
		}
		context.Apps[position].Models = append(context.Apps[position].Models, IndexModel{
			AppLabel:  appLabel,
			Name:      admin.Model.ModelName,
			AddURL:    addURL,
			ChangeURL: changeURL,
		})
	}
	return context, nil
}

func canViewOrChange(user auth.User, appLabel, modelName string) bool {
	return auth.HasPerm(user, appLabel+".view_"+modelName) || auth.HasPerm(user, appLabel+".change_"+modelName)
}

func routeName(appLabel, modelName, action string) string {
	return "admin:" + appLabel + "_" + modelName + "_" + action
}
