export namespace main {
	
	export class AthleteDTO {
	    id: string;
	    categoryId: string;
	    name: string;
	    club: string;
	    weight: number;
	    birthDate: string;
	
	    static createFrom(source: any = {}) {
	        return new AthleteDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.categoryId = source["categoryId"];
	        this.name = source["name"];
	        this.club = source["club"];
	        this.weight = source["weight"];
	        this.birthDate = source["birthDate"];
	    }
	}
	export class MatchDTO {
	    id: string;
	    round: number;
	    position: number;
	    athleteA?: AthleteDTO;
	    athleteB?: AthleteDTO;
	    status: string;
	    tatamiId: string;
	    isRepechage: boolean;
	    winnerId?: string;
	    method?: string;
	
	    static createFrom(source: any = {}) {
	        return new MatchDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.round = source["round"];
	        this.position = source["position"];
	        this.athleteA = this.convertValues(source["athleteA"], AthleteDTO);
	        this.athleteB = this.convertValues(source["athleteB"], AthleteDTO);
	        this.status = source["status"];
	        this.tatamiId = source["tatamiId"];
	        this.isRepechage = source["isRepechage"];
	        this.winnerId = source["winnerId"];
	        this.method = source["method"];
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
	export class RoundDTO {
	    number: number;
	    matches: MatchDTO[];
	
	    static createFrom(source: any = {}) {
	        return new RoundDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.number = source["number"];
	        this.matches = this.convertValues(source["matches"], MatchDTO);
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
	export class BracketDTO {
	    categoryId: string;
	    rounds: RoundDTO[];
	    repechage: RoundDTO[];
	
	    static createFrom(source: any = {}) {
	        return new BracketDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.categoryId = source["categoryId"];
	        this.rounds = this.convertValues(source["rounds"], RoundDTO);
	        this.repechage = this.convertValues(source["repechage"], RoundDTO);
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
	export class DivisionDTO {
	    id: string;
	    tournamentId: string;
	    ageGroup: string;
	    gender: string;
	    weightClass: string;
	    format: string;
	
	    static createFrom(source: any = {}) {
	        return new DivisionDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.tournamentId = source["tournamentId"];
	        this.ageGroup = source["ageGroup"];
	        this.gender = source["gender"];
	        this.weightClass = source["weightClass"];
	        this.format = source["format"];
	    }
	}
	
	
	export class TournamentDTO {
	    id: string;
	    name: string;
	    location: string;
	    date: string;
	
	    static createFrom(source: any = {}) {
	        return new TournamentDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.location = source["location"];
	        this.date = source["date"];
	    }
	}

}

export namespace ports {
	
	export class MatchRow {
	    ID: string;
	    CategoryID: string;
	    DivisionID: string;
	    WeightClass: string;
	    Gender: string;
	    AgeGroup: string;
	    Round: number;
	    Position: number;
	    IsRepechage: boolean;
	    AthleteAID: string;
	    AthleteBID: string;
	    AthleteAName: string;
	    AthleteAClub: string;
	    AthleteBName: string;
	    AthleteBClub: string;
	    Status: string;
	    TatamiID: string;
	    ResultJSON: string;
	
	    static createFrom(source: any = {}) {
	        return new MatchRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CategoryID = source["CategoryID"];
	        this.DivisionID = source["DivisionID"];
	        this.WeightClass = source["WeightClass"];
	        this.Gender = source["Gender"];
	        this.AgeGroup = source["AgeGroup"];
	        this.Round = source["Round"];
	        this.Position = source["Position"];
	        this.IsRepechage = source["IsRepechage"];
	        this.AthleteAID = source["AthleteAID"];
	        this.AthleteBID = source["AthleteBID"];
	        this.AthleteAName = source["AthleteAName"];
	        this.AthleteAClub = source["AthleteAClub"];
	        this.AthleteBName = source["AthleteBName"];
	        this.AthleteBClub = source["AthleteBClub"];
	        this.Status = source["Status"];
	        this.TatamiID = source["TatamiID"];
	        this.ResultJSON = source["ResultJSON"];
	    }
	}

}

