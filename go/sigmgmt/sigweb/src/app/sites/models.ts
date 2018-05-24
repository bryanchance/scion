export class Site {
    ID: number
    Name = ''
    VHost: string
    Hosts: Host[] = []
    MetricsPort: number
    constructor(Name: string = '') {
        this.Name = Name
    }
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

export class PathSelector {
    ID: string
    Name: string
    Filter: string
}

export class ASEntry {
    ID: number
    Name: string
    ISD: string
    AS: string
    Policies: string
}

export class Policy {
    Policy: string
    constructor(policy: string) {
        this.Policy = policy
    }
}

export class CIDR {
    ID: number
    CIDR: string
}

export class SIG {
    ID: number
    Name: string
    Address: string
    EncapPort: number
    CtrlPort: number
}
