{
	"pathPattern": "/containers/create",
	"function" : function(ctx, plugins){
		ctx.body.Labels["test_label"] = ctx.body.Labels["test_label"] + "_edited_by_filter_set_var";
		ctx.values.set("label", ctx.body.Labels["test_label"]); 
		return {status: 200, next: true, body: ctx.body, operation : ctx.operation.WRITE};
	}
}

