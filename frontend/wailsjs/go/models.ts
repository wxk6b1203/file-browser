export namespace render {
	
	export class DragSignal {
	    type: string;
	    x: number;
	    y: number;
	    paths?: string[];
	
	    static createFrom(source: any = {}) {
	        return new DragSignal(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.paths = source["paths"];
	    }
	}
	export class NodeRect {
	    x: number;
	    y: number;
	    width: number;
	    height: number;
	
	    static createFrom(source: any = {}) {
	        return new NodeRect(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	        this.width = source["width"];
	        this.height = source["height"];
	    }
	}

}

