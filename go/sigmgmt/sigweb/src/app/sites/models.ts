export class Site {
    Name = ''
    VHost: string
    Hosts: Host[] = []
    MetricsPort: number
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

export class Session {
    ID: number
    FilterName: string // PathSelector name
}
