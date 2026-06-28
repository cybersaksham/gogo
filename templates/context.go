package templates

// Context is the template render context.
type Context map[string]any

// ContextProcessor injects request-scoped values before rendering.
type ContextProcessor func(Context) Context

// Message is a simple request message exposed to templates.
type Message struct {
	Level string
	Text  string
}

func RequestProcessor(request any) ContextProcessor {
	return func(context Context) Context {
		context["request"] = request
		return context
	}
}

func UserProcessor(user any) ContextProcessor {
	return func(context Context) Context {
		context["user"] = user
		return context
	}
}

func MessagesProcessor(messages []Message) ContextProcessor {
	return func(context Context) Context {
		context["messages"] = append([]Message(nil), messages...)
		return context
	}
}

func CSRFTokenProcessor(token string) ContextProcessor {
	return func(context Context) Context {
		context["csrf_token"] = token
		return context
	}
}

func StaticURLProcessor(url string) ContextProcessor {
	return func(context Context) Context {
		context["static_url"] = url
		return context
	}
}

func MediaURLProcessor(url string) ContextProcessor {
	return func(context Context) Context {
		context["media_url"] = url
		return context
	}
}

func cloneContext(context Context) Context {
	cloned := Context{}
	for key, value := range context {
		cloned[key] = value
	}
	return cloned
}
