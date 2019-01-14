{
	"pathPattern": "/containers/create",
	"function" : function(ctx, plugins){
		if (ctx.values.get("label") != "go-horse-label_edited_by_filter_set_var") {
			return {status: 500, next: false, body: ctx.body, operation : ctx.operation.READ};
		}
		return {status: 200, next: true, body: ctx.body, operation : ctx.operation.READ};
	}
}

