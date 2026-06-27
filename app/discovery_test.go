package app

import "testing"

func TestRegistryDiscoveryMethodsReturnCopies(t *testing.T) {
	registry := NewRegistry()

	registry.RegisterModel(ModelResource{AppLabel: "blog", Name: "Post"})
	registry.RegisterAdmin(AdminResource{AppLabel: "blog", ModelName: "Post", Handler: "PostAdmin"})
	registry.RegisterRoute(RouteResource{AppLabel: "blog", Name: "blog:index", Path: "/blog/", Handler: "Index"})
	registry.RegisterAPIRoute(APIRouteResource{AppLabel: "blog", Name: "blog-api:list", Path: "/api/posts/", Handler: "PostViewSet"})
	registry.RegisterForm(FormResource{AppLabel: "blog", Name: "PostForm", Handler: "PostForm"})
	registry.RegisterTemplate(TemplateResource{AppLabel: "blog", Path: "blog/post.html"})
	registry.RegisterStaticRoot(StaticResource{AppLabel: "blog", Path: "blog/static"})
	registry.RegisterTask(TaskResource{AppLabel: "blog", Name: "blog.publish", Handler: "Publish"})
	registry.RegisterCommand(CommandResource{AppLabel: "blog", Name: "blog.reindex", Handler: "Reindex"})
	registry.RegisterMigration(MigrationResource{AppLabel: "blog", Name: "0001_initial"})

	models := registry.Models()
	models[0].Name = "Changed"
	if got := registry.Models()[0].Name; got != "Post" {
		t.Fatalf("Models() leaked internal slice, got %q", got)
	}

	admin := registry.Admin()
	admin[0].Handler = "Changed"
	if got := registry.Admin()[0].Handler; got != "PostAdmin" {
		t.Fatalf("Admin() leaked internal slice, got %q", got)
	}

	routes := registry.Routes()
	routes[0].Path = "/changed/"
	if got := registry.Routes()[0].Path; got != "/blog/" {
		t.Fatalf("Routes() leaked internal slice, got %q", got)
	}

	apiRoutes := registry.APIRoutes()
	apiRoutes[0].Path = "/changed/"
	if got := registry.APIRoutes()[0].Path; got != "/api/posts/" {
		t.Fatalf("APIRoutes() leaked internal slice, got %q", got)
	}

	forms := registry.Forms()
	forms[0].Handler = "Changed"
	if got := registry.Forms()[0].Handler; got != "PostForm" {
		t.Fatalf("Forms() leaked internal slice, got %q", got)
	}

	templates := registry.Templates()
	templates[0].Path = "changed.html"
	if got := registry.Templates()[0].Path; got != "blog/post.html" {
		t.Fatalf("Templates() leaked internal slice, got %q", got)
	}

	staticRoots := registry.StaticRoots()
	staticRoots[0].Path = "changed"
	if got := registry.StaticRoots()[0].Path; got != "blog/static" {
		t.Fatalf("StaticRoots() leaked internal slice, got %q", got)
	}

	tasks := registry.Tasks()
	tasks[0].Handler = "Changed"
	if got := registry.Tasks()[0].Handler; got != "Publish" {
		t.Fatalf("Tasks() leaked internal slice, got %q", got)
	}

	commands := registry.Commands()
	commands[0].Handler = "Changed"
	if got := registry.Commands()[0].Handler; got != "Reindex" {
		t.Fatalf("Commands() leaked internal slice, got %q", got)
	}

	migrations := registry.Migrations()
	migrations[0].Name = "changed"
	if got := registry.Migrations()[0].Name; got != "0001_initial" {
		t.Fatalf("Migrations() leaked internal slice, got %q", got)
	}
}
