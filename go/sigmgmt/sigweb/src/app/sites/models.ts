export class Site {
    Name = ''
    VHost: string
    Hosts: Host[] = []
    MetricsPort: number
    constructor(Name: string = '') {
        this.Name = Name
    }
}

export class Host {
    Name: string
    User: string
    Key: string
    constructor() {
        this.User = ''
        this.Key = ''
    }
}

export class PathSelector {
    Name: string
    PP: string
}

export class IA {
    Name: string
    ISD: string
    AS: string
    Policy: string
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
    ID: string
    Addr: string
    EncapPort: number
    CtrlPort: number
}
