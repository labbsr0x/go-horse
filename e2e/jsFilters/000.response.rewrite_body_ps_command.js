{
	"pathPattern": "containers/json",
	"function" : function(ctx, plugins){
		ctx.body.push({Command:"go-horse.sh",Created:1547217151,HostConfig:{NetworkMode:"default"},Id:"5a3c9bbfc048c3c84d066b23883595f7c7d5b1d4055a94fd8bfd055be382b681",Image:"go-horse-image",ImageID:"sha256:415381a6cb813ef0972eff8edac32069637b4546349d9ffdb8e4f641f55edcdd",Labels:{},Mounts:[{Destination:"/data",Driver:"local",Mode:"",Name:"cf5c681ebd72480485b41c38519b0d8eeaea38946c60a5b73f3cdc15901ef436",Propagation:"",RW:true,Source:"",Type:"volume"}],Names:["/go-horse-name"],NetworkSettings:{Networks:{bridge:{Aliases:null,DriverOpts:null,EndpointID:"d885277b530d89d8857478f843383e69e3c03f94a3b89341b3e49f7f9dcf6754",Gateway:"192.168.1.5",GlobalIPv6Address:"",GlobalIPv6PrefixLen:0,IPAMConfig:null,IPAddress:"192.168.1.1",IPPrefixLen:24,IPv6Gateway:"",Links:null,MacAddress:"02:42:c0:a8:01:01",NetworkID:"ed27491b26dced3278b76d5bd604ddad7f1300455c92e17833f625b07f1ea778"}}},Ports:[{IP:"0.0.0.0",PrivatePort:6379,PublicPort:6379,Type:"tcp"}],State:"go-horsing",Status:"About a go-horse ago"});
		return {status: 200, next: true, body: ctx.body, operation : ctx.operation.WRITE};
	}
}

