export namespace config {
	
	export class UISection {
	    locale: string;
	    theme: string;
	    explorerFontSize: number;
	    fileListFontSize: number;
	
	    static createFrom(source: any = {}) {
	        return new UISection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.locale = source["locale"];
	        this.theme = source["theme"];
	        this.explorerFontSize = source["explorerFontSize"];
	        this.fileListFontSize = source["fileListFontSize"];
	    }
	}
	export class TransferSection {
	    tempDir: string;
	    downloadDir: string;
	    overwriteStrategy: string;
	
	    static createFrom(source: any = {}) {
	        return new TransferSection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tempDir = source["tempDir"];
	        this.downloadDir = source["downloadDir"];
	        this.overwriteStrategy = source["overwriteStrategy"];
	    }
	}
	export class SearchSection {
	    maxConcurrency: number;
	    resultLimit: number;
	
	    static createFrom(source: any = {}) {
	        return new SearchSection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.maxConcurrency = source["maxConcurrency"];
	        this.resultLimit = source["resultLimit"];
	    }
	}
	export class PathSection {
	    connectionsFile: string;
	    stateFile: string;
	
	    static createFrom(source: any = {}) {
	        return new PathSection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectionsFile = source["connectionsFile"];
	        this.stateFile = source["stateFile"];
	    }
	}
	export class LogSection {
	    level: string;
	    outputs: string[];
	
	    static createFrom(source: any = {}) {
	        return new LogSection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.level = source["level"];
	        this.outputs = source["outputs"];
	    }
	}
	export class AppSection {
	    locale: string;
	    theme: string;
	    tempDir: string;
	
	    static createFrom(source: any = {}) {
	        return new AppSection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.locale = source["locale"];
	        this.theme = source["theme"];
	        this.tempDir = source["tempDir"];
	    }
	}
	export class AppConfig {
	    app: AppSection;
	    log: LogSection;
	    paths: PathSection;
	    search: SearchSection;
	    transfer: TransferSection;
	    ui: UISection;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.app = this.convertValues(source["app"], AppSection);
	        this.log = this.convertValues(source["log"], LogSection);
	        this.paths = this.convertValues(source["paths"], PathSection);
	        this.search = this.convertValues(source["search"], SearchSection);
	        this.transfer = this.convertValues(source["transfer"], TransferSection);
	        this.ui = this.convertValues(source["ui"], UISection);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	

}

export namespace connection {
	
	export class Definition {
	    id: string;
	    name: string;
	    driver: string;
	    description?: string;
	    enabled: boolean;
	    readOnly?: boolean;
	    root?: string;
	    tags?: string[];
	    metadata?: Record<string, string>;
	    config?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new Definition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.driver = source["driver"];
	        this.description = source["description"];
	        this.enabled = source["enabled"];
	        this.readOnly = source["readOnly"];
	        this.root = source["root"];
	        this.tags = source["tags"];
	        this.metadata = source["metadata"];
	        this.config = source["config"];
	    }
	}
	export class State {
	    id: string;
	    name: string;
	    driver: string;
	    connected: boolean;
	    lastError?: string;
	    // Go type: time
	    lastConnectedAt?: any;
	    capabilities?: folder.Capabilities;
	
	    static createFrom(source: any = {}) {
	        return new State(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.driver = source["driver"];
	        this.connected = source["connected"];
	        this.lastError = source["lastError"];
	        this.lastConnectedAt = this.convertValues(source["lastConnectedAt"], null);
	        this.capabilities = this.convertValues(source["capabilities"], folder.Capabilities);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace folder {
	
	export class Capabilities {
	    CanList: boolean;
	    CanRead: boolean;
	    CanWrite: boolean;
	    CanDelete: boolean;
	    CanCopy: boolean;
	    CanMove: boolean;
	    CanRename: boolean;
	    CanMkdir: boolean;
	    CanPresign: boolean;
	    CanTransfer: boolean;
	    AtomicMove: boolean;
	    SupportsVersion: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Capabilities(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.CanList = source["CanList"];
	        this.CanRead = source["CanRead"];
	        this.CanWrite = source["CanWrite"];
	        this.CanDelete = source["CanDelete"];
	        this.CanCopy = source["CanCopy"];
	        this.CanMove = source["CanMove"];
	        this.CanRename = source["CanRename"];
	        this.CanMkdir = source["CanMkdir"];
	        this.CanPresign = source["CanPresign"];
	        this.CanTransfer = source["CanTransfer"];
	        this.AtomicMove = source["AtomicMove"];
	        this.SupportsVersion = source["SupportsVersion"];
	    }
	}
	export class DriverInfo {
	    name: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new DriverInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	    }
	}
	export class Owner {
	    id: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new Owner(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}
	export class FileInfo {
	    name: string;
	    path: string;
	    type: number;
	    size: number;
	    // Go type: time
	    lastModified?: any;
	    owner?: Owner;
	    contentType?: string;
	    etag?: string;
	    metadata?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.type = source["type"];
	        this.size = source["size"];
	        this.lastModified = this.convertValues(source["lastModified"], null);
	        this.owner = this.convertValues(source["owner"], Owner);
	        this.contentType = source["contentType"];
	        this.etag = source["etag"];
	        this.metadata = source["metadata"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class TransferTask {
	    id: string;
	    driverName: string;
	    instanceName: string;
	    direction: number;
	    remotePath: string;
	    localPath: string;
	    status: number;
	    bytesTransferred: number;
	    totalBytes: number;
	    bytesPerSecond: number;
	    error?: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    startedAt?: any;
	    // Go type: time
	    completedAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new TransferTask(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.driverName = source["driverName"];
	        this.instanceName = source["instanceName"];
	        this.direction = source["direction"];
	        this.remotePath = source["remotePath"];
	        this.localPath = source["localPath"];
	        this.status = source["status"];
	        this.bytesTransferred = source["bytesTransferred"];
	        this.totalBytes = source["totalBytes"];
	        this.bytesPerSecond = source["bytesPerSecond"];
	        this.error = source["error"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.startedAt = this.convertValues(source["startedAt"], null);
	        this.completedAt = this.convertValues(source["completedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

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
	export class PanelDropSignal {
	    groupId: string;
	    tabId: string;
	    x: number;
	    y: number;
	    paths?: string[];
	
	    static createFrom(source: any = {}) {
	        return new PanelDropSignal(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.groupId = source["groupId"];
	        this.tabId = source["tabId"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.paths = source["paths"];
	    }
	}

}

export namespace search {
	
	export class Request {
	    query: string;
	    connectionIds?: string[];
	    root?: string;
	    maxResults?: number;
	
	    static createFrom(source: any = {}) {
	        return new Request(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.query = source["query"];
	        this.connectionIds = source["connectionIds"];
	        this.root = source["root"];
	        this.maxResults = source["maxResults"];
	    }
	}

}

