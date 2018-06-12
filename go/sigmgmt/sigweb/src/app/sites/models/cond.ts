import { TrafficClass } from './models'
export const TypeCondAllOf = 'CondAllOf'
export const TypeCondAnyOf = 'CondAnyOf'
export const TypeCondNot = 'CondNot'
export const TypeCondIPv4 = 'CondIPv4'
export const TypeCondClass = 'CondClass'
export const TypeIPv4MatchSource = 'MatchSource'
export const TypeIPv4MatchDestination = 'MatchDestination'
export const TypeIPv4MatchDSCP = 'MatchDSCP'

export interface Cond {
    Type(): string
    String(): string
}

export interface IPv4Predicate {
    Type(): string
    String(): string
}

export function unmarshalInterface(json): Cond {
    if (json[TypeCondAllOf]) {
        return CondAllFromJSON(json[TypeCondAllOf])
    }
    if (json[TypeCondAnyOf]) {
        return CondAnyFromJSON(json[TypeCondAnyOf])
    }
    if (json[TypeCondNot]) {
        return CondNotFromJSON(json[TypeCondNot])
    }
    if (json[TypeCondIPv4]) {
        return CondIPv4FromJSON(json[TypeCondIPv4])
    }
    if (json[TypeCondClass]) {
        return CondClassFromJSON(json[TypeCondClass])
    }
    if (json[TypeIPv4MatchSource]) {
        return MatchSourceFromJSON(json[TypeIPv4MatchSource])
    }
    if (json[TypeIPv4MatchDestination]) {
        return MatchDestinationFromJSON(json[TypeIPv4MatchDestination])
    }
    if (json[TypeIPv4MatchDSCP]) {
        return MatchdscpFromJSON(json[TypeIPv4MatchDSCP])
    }
    console.error(json)
}

export type CondAll = Cond[]
export class CondAllOf {
    Conds: CondAll = []

    Type() {
        return TypeCondAllOf
    }

    String() {
        let s = 'ALL('
        for (const cond of this.Conds) {
            s += cond.String() + ','
        }
        if (this.Conds.length > 0) {
            s = s.slice(0, -1)
        }
        s += ')'
        return s
    }
}

function CondAllFromJSON(json: CondAll): CondAllOf {
    const condAll = new CondAllOf()
    for (const cond of json) {
        condAll.Conds.push(unmarshalInterface(cond))
    }
    return condAll
}

export type CondAny = Cond[]
export class CondAnyOf {
    Conds: CondAny = []

    Type() {
        return TypeCondAnyOf
    }

    String() {
        let s = 'ANY('
        for (const cond of this.Conds) {
            s += cond.String() + ','
        }
        if (this.Conds.length > 0) {
            s = s.slice(0, -1)
        }
        s += ')'
        return s
    }
}

function CondAnyFromJSON(json: CondAny): CondAnyOf {
    const condAny = new CondAnyOf()
    for (const cond of json) {
        condAny.Conds.push(unmarshalInterface(cond))
    }
    return condAny
}

export class CondNot implements Cond {
    Operand: Cond

    Type() {
        return TypeCondNot
    }

    String() {
        return this.Operand ? 'NOT(' + this.Operand.String() + ')' : ''
    }
}

function CondNotFromJSON(json: CondNot): CondNot {
    const condNot = new CondNot()
    condNot.Operand = unmarshalInterface(json)
    return condNot
}

export class CondIPv4 implements Cond {
    Predicate: IPv4Predicate

    Type() {
        return TypeCondIPv4
    }

    String() {
        return this.Predicate.String()
    }
}

function CondIPv4FromJSON(json: CondIPv4): CondIPv4 {
    const condIPv4 = new CondIPv4()
    condIPv4.Predicate = unmarshalInterface(json)
    return condIPv4
}

export class CondClass implements Cond {
    TrafficClass: number

    Type() {
        return TypeCondClass
    }

    String() {
        return this.TrafficClass ? 'cls=' + this.TrafficClass : 'cls='
    }
}

function CondClassFromJSON(json: CondClass): CondClass {
    const condClass = new CondClass()
    condClass.TrafficClass = +json.TrafficClass
    return condClass
}


export class MatchSource implements IPv4Predicate {
    Net: string

    Type() {
        return TypeIPv4MatchSource
    }

    String() {
        return this.Net ? 'src=' + this.Net : 'src='
    }
}

function MatchSourceFromJSON(json: MatchSource): MatchSource {
    const msrc = new MatchSource()
    msrc.Net = json.Net
    return msrc
}

export class MatchDestination implements IPv4Predicate {
    Net: string

    Type() {
        return TypeIPv4MatchDestination
    }

    String() {
        return this.Net ? 'dst=' + this.Net : 'dst='
    }
}

function MatchDestinationFromJSON(json: MatchDestination): MatchDestination {
    const mdst = new MatchDestination()
    mdst.Net = json.Net
    return mdst
}

export class MatchDSCP implements IPv4Predicate {
    DSCP: number

    Type() {
        return TypeIPv4MatchDSCP
    }

    String() {
        return this.DSCP ? 'dscp=' + this.DSCP : 'dscp='
    }
}

function MatchdscpFromJSON(json): MatchDSCP {
    const dscp = new MatchDSCP()
    dscp.DSCP = json.DSCP
    return dscp
}
