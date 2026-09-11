export namespace config {
	
	export class Config {
	    refreshInterval: number;
	    startWithWindows: boolean;
	    hideOnStartup: boolean;
	    showHoverPopup: boolean;
	    trayMetric: string;
	    showLongBar: boolean;
	    notify5hBelow: number;
	    notify7dBelow: number;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.refreshInterval = source["refreshInterval"];
	        this.startWithWindows = source["startWithWindows"];
	        this.hideOnStartup = source["hideOnStartup"];
	        this.showHoverPopup = source["showHoverPopup"];
	        this.trayMetric = source["trayMetric"];
	        this.showLongBar = source["showLongBar"];
	        this.notify5hBelow = source["notify5hBelow"];
	        this.notify7dBelow = source["notify7dBelow"];
	    }
	}

}

export namespace usage {
	
	export class CreditsInfo {
	    hasCredits: boolean;
	    unlimited: boolean;
	    balance: string;
	
	    static createFrom(source: any = {}) {
	        return new CreditsInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hasCredits = source["hasCredits"];
	        this.unlimited = source["unlimited"];
	        this.balance = source["balance"];
	    }
	}
	export class QuotaWindow {
	    usedPct: number;
	    remainingPct: number;
	    windowMinutes: number;
	    resetsAt?: number;
	
	    static createFrom(source: any = {}) {
	        return new QuotaWindow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.usedPct = source["usedPct"];
	        this.remainingPct = source["remainingPct"];
	        this.windowMinutes = source["windowMinutes"];
	        this.resetsAt = source["resetsAt"];
	    }
	}
	export class TokenStats {
	    today: number;
	    near7d: number;
	    total: number;
	    currentTask: number;
	    lastTurn: number;
	    cacheHitRate: number;
	
	    static createFrom(source: any = {}) {
	        return new TokenStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.today = source["today"];
	        this.near7d = source["near7d"];
	        this.total = source["total"];
	        this.currentTask = source["currentTask"];
	        this.lastTurn = source["lastTurn"];
	        this.cacheHitRate = source["cacheHitRate"];
	    }
	}
	export class Snapshot {
	    id: string;
	    name: string;
	    shortName: string;
	    color: string;
	    emoji: string;
	    primary?: QuotaWindow;
	    secondary?: QuotaWindow;
	    token: TokenStats;
	    credits?: CreditsInfo;
	    status: string;
	    source: string;
	    quotaSource: string;
	    lastRequestAt: number;
	    ts: number;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new Snapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.shortName = source["shortName"];
	        this.color = source["color"];
	        this.emoji = source["emoji"];
	        this.primary = this.convertValues(source["primary"], QuotaWindow);
	        this.secondary = this.convertValues(source["secondary"], QuotaWindow);
	        this.token = this.convertValues(source["token"], TokenStats);
	        this.credits = this.convertValues(source["credits"], CreditsInfo);
	        this.status = source["status"];
	        this.source = source["source"];
	        this.quotaSource = source["quotaSource"];
	        this.lastRequestAt = source["lastRequestAt"];
	        this.ts = source["ts"];
	        this.error = source["error"];
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

