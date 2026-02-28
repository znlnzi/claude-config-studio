export namespace services {
	
	export class ExtensionFile {
	    name: string;
	    fileName: string;
	    path: string;
	    content: string;
	    lastModified: string;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new ExtensionFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.fileName = source["fileName"];
	        this.path = source["path"];
	        this.content = source["content"];
	        this.lastModified = source["lastModified"];
	        this.size = source["size"];
	    }
	}
	export class GlobalConfig {
	    claudeHome: string;
	    claudeMd: string;
	    settings: number[];
	    hasClaudeMd: boolean;
	    hasSettings: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GlobalConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.claudeHome = source["claudeHome"];
	        this.claudeMd = source["claudeMd"];
	        this.settings = source["settings"];
	        this.hasClaudeMd = source["hasClaudeMd"];
	        this.hasSettings = source["hasSettings"];
	    }
	}
	export class HookCommand {
	    type: string;
	    command: string;
	    timeout?: number;
	
	    static createFrom(source: any = {}) {
	        return new HookCommand(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.command = source["command"];
	        this.timeout = source["timeout"];
	    }
	}
	export class HookEntry {
	    matcher?: string;
	    hooks: HookCommand[];
	
	    static createFrom(source: any = {}) {
	        return new HookEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.matcher = source["matcher"];
	        this.hooks = this.convertValues(source["hooks"], HookCommand);
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
	export class HooksConfig {
	    event: string;
	    entries: HookEntry[];
	
	    static createFrom(source: any = {}) {
	        return new HooksConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.event = source["event"];
	        this.entries = this.convertValues(source["entries"], HookEntry);
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
	export class MCPServer {
	    name: string;
	    type?: string;
	    url?: string;
	    headers?: Record<string, string>;
	    command?: string;
	    args?: string[];
	    timeout?: number;
	
	    static createFrom(source: any = {}) {
	        return new MCPServer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.url = source["url"];
	        this.headers = source["headers"];
	        this.command = source["command"];
	        this.args = source["args"];
	        this.timeout = source["timeout"];
	    }
	}
	export class MarketplaceServer {
	    name: string;
	    description: string;
	    descriptionCN: string;
	    repoUrl: string;
	    package: string;
	    transport: string;
	    command: string;
	    args: string[];
	    version: string;
	    publishedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new MarketplaceServer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.descriptionCN = source["descriptionCN"];
	        this.repoUrl = source["repoUrl"];
	        this.package = source["package"];
	        this.transport = source["transport"];
	        this.command = source["command"];
	        this.args = source["args"];
	        this.version = source["version"];
	        this.publishedAt = source["publishedAt"];
	    }
	}
	export class MarketplaceResult {
	    servers: MarketplaceServer[];
	    nextCursor: string;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new MarketplaceResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.servers = this.convertValues(source["servers"], MarketplaceServer);
	        this.nextCursor = source["nextCursor"];
	        this.total = source["total"];
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
	
	export class OnlineExtension {
	    name: string;
	    description: string;
	    category: string;
	    source: string;
	    repoUrl: string;
	    downloadUrl: string;
	    extType: string;
	
	    static createFrom(source: any = {}) {
	        return new OnlineExtension(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.source = source["source"];
	        this.repoUrl = source["repoUrl"];
	        this.downloadUrl = source["downloadUrl"];
	        this.extType = source["extType"];
	    }
	}
	export class OnlineExtensionResult {
	    extensions: OnlineExtension[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new OnlineExtensionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.extensions = this.convertValues(source["extensions"], OnlineExtension);
	        this.total = source["total"];
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
	export class PluginInfo {
	    name: string;
	    source: string;
	    enabled: boolean;
	    description?: string;
	    version?: string;
	
	    static createFrom(source: any = {}) {
	        return new PluginInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.source = source["source"];
	        this.enabled = source["enabled"];
	        this.description = source["description"];
	        this.version = source["version"];
	    }
	}
	export class ProjectConfig {
	    path: string;
	    claudeMd: string;
	    settings: number[];
	    mcpConfig: number[];
	    hasClaudeMd: boolean;
	    hasSettings: boolean;
	    hasMcp: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProjectConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.claudeMd = source["claudeMd"];
	        this.settings = source["settings"];
	        this.mcpConfig = source["mcpConfig"];
	        this.hasClaudeMd = source["hasClaudeMd"];
	        this.hasSettings = source["hasSettings"];
	        this.hasMcp = source["hasMcp"];
	    }
	}
	export class ProjectInfo {
	    name: string;
	    path: string;
	    hasClaudeMd: boolean;
	    hasSettings: boolean;
	    hasMcp: boolean;
	    hasHooks: boolean;
	    hasCommands: boolean;
	    hasAgents: boolean;
	    hasSkills: boolean;
	    // Go type: time
	    lastModified: any;
	    configCount: number;
	
	    static createFrom(source: any = {}) {
	        return new ProjectInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.hasClaudeMd = source["hasClaudeMd"];
	        this.hasSettings = source["hasSettings"];
	        this.hasMcp = source["hasMcp"];
	        this.hasHooks = source["hasHooks"];
	        this.hasCommands = source["hasCommands"];
	        this.hasAgents = source["hasAgents"];
	        this.hasSkills = source["hasSkills"];
	        this.lastModified = this.convertValues(source["lastModified"], null);
	        this.configCount = source["configCount"];
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
	export class SkillFileInfo {
	    relativePath: string;
	    isDir: boolean;
	    size: number;
	    isMain: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SkillFileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.relativePath = source["relativePath"];
	        this.isDir = source["isDir"];
	        this.size = source["size"];
	        this.isMain = source["isMain"];
	    }
	}
	export class SkillInfo {
	    name: string;
	    description: string;
	    source: string;
	    marketplace: string;
	    pluginName: string;
	    type: string;
	    filePath: string;
	
	    static createFrom(source: any = {}) {
	        return new SkillInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.source = source["source"];
	        this.marketplace = source["marketplace"];
	        this.pluginName = source["pluginName"];
	        this.type = source["type"];
	        this.filePath = source["filePath"];
	    }
	}
	export class UserSkillInfo {
	    name: string;
	    description: string;
	    scope: string;
	    dirName: string;
	    isFlat: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UserSkillInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.scope = source["scope"];
	        this.dirName = source["dirName"];
	        this.isFlat = source["isFlat"];
	    }
	}

}

export namespace templatedata {
	
	export class InstalledTemplateInfo {
	    templateId: string;
	    scope: string;
	    filePath: string;
	
	    static createFrom(source: any = {}) {
	        return new InstalledTemplateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.templateId = source["templateId"];
	        this.scope = source["scope"];
	        this.filePath = source["filePath"];
	    }
	}
	export class Template {
	    id: string;
	    name: string;
	    category: string;
	    description: string;
	    tags: string[];
	    claudeMd?: string;
	    settings?: any;
	    mcpServers?: any;
	    hooks?: any;
	    agents?: Record<string, string>;
	    commands?: Record<string, string>;
	    skills?: Record<string, string>;
	    rules?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Template(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.category = source["category"];
	        this.description = source["description"];
	        this.tags = source["tags"];
	        this.claudeMd = source["claudeMd"];
	        this.settings = source["settings"];
	        this.mcpServers = source["mcpServers"];
	        this.hooks = source["hooks"];
	        this.agents = source["agents"];
	        this.commands = source["commands"];
	        this.skills = source["skills"];
	        this.rules = source["rules"];
	    }
	}
	export class TemplateCategory {
	    id: string;
	    name: string;
	    icon: string;
	    templates: Template[];
	
	    static createFrom(source: any = {}) {
	        return new TemplateCategory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.icon = source["icon"];
	        this.templates = this.convertValues(source["templates"], Template);
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

