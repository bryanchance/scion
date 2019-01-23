import { Cond, unmarshalInterface } from './cond'

export class Site {
    ID: number
    Name = ''
    VHost: string
    Hosts: Host[] = []
    MetricsPort: number
    constructor(ID) {
        this.ID = ID
    }
}

export class SiteNetwork {
    ID: number
    CIDR: string
    ACL: string
}

export class Host {
    ID: number
    Name: string
    User: string
    Key: string
    constructor() {
        this.User = ''
        this.Key = ''
    }
}

export class PathPolicyFile {
    ID: number
    Name: string
    Code: object[]
    Type = 'global'
    SiteID: number
}

export class PathPolicy {
    constructor(public Name: string, public Policy) { }
}

export class ASEntry {
    ID: number
    Name: string
    ISD: string
    AS: string
    IPAllocationProvider: string
}

export class TrafficPolicy {
    ID: number
    Name: string
    TrafficClass: number
    PathPolicies: string[] = []
}

export class CIDR {
    ID: number
    CIDR: string
    Scraped: boolean
}

export class TrafficClass {
    ID: number
    Name: string
    Cond: Cond
    CondStr: string

    get condString(): string {
        return this.Cond ? this.Cond.String() : ''
    }
}

export function TrafficClassFromJSON(json: TrafficClass): TrafficClass {
    const trClass = new TrafficClass()
    trClass.ID = json.ID
    trClass.Name = json.Name
    trClass.Cond = unmarshalInterface(json.Cond)
    trClass.CondStr = ''
    return trClass
}
