export namespace shortcut {
	
	export class Shortcut {
	    id: string;
	    label: string;
	    accelerator: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Shortcut(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.accelerator = source["accelerator"];
	        this.enabled = source["enabled"];
	    }
	}

}

