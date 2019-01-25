#!/usr/bin/env python3

import yaml
from tiles import t

with open("schema.yml", 'r') as f:
    sch = yaml.load(f)

fields = (t/'').vjoin([t/'@{e} map[string]*@{e}' for e in sch])

out = t/"""
    // This file was generated from schema.yml by schema/generator

    package schema

    import "github.com/scionproto/scion/go/lib/common"

    type Layout struct {
        Generator string
        GeneratorVersion string
        GeneratorBuildChain string
        @{fields}
    }
    """

for e, entity in sch.items():
    fields = t/''
    functs = t/''
    for f, field in entity.items():
        if field["type"] == "Ref":
            fields |= t/"""
                @{f} *@{field["refentity"]} `toml:"-"`
                REF@{f} string `toml:"@{f},omitempty"`
                """
            functs |= t%'' | t/"""
                func (self *@{e}) Set@{f}(ref *@{field["refentity"]}) error {
                    if self.@{f} != nil {
                        return common.NewBasicError("Reference is already set", nil)
                    }
                    self._@{f}(ref)
                    ref._@{field["reffield"]}(self)
                    return nil
                }

                func (self *@{e}) _@{f}(ref *@{field["refentity"]}) {
                    self.@{f} = ref
                }
                """
        elif field["type"] == "Refs":
            fields |= t/"""
                @{f} []*@{field["refentity"]} `toml:"-"`
                REF@{f} []string `toml:"@{f},omitempty"`
                """
            functs |= t%'' | t/"""
                func (self *@{e}) Add@{f}(ref *@{field["refentity"]}) error {
                    for _, r := range self.@{f} {
                        if r == ref {
                            return common.NewBasicError("Reference is already present", nil)
                        }
                    }
                    self._@{f}(ref)
                    ref._@{field["reffield"]}(self)
                    return nil
                }

                func (self *@{e}) _@{f}(ref *@{field["refentity"]}) {
                    if self.@{f} == nil {
                        self.@{f} = make([]*@{field["refentity"]}, 0)
                    }
                    self.@{f} = append(self.@{f}, ref)
                }
                """
        else:
            fields |= t/"""
                @{f} @{field["type"]} `toml:"@{f},omitempty"`
                """

    out |= t%'' | t/"""
        // @{e} object.

        type @{e} struct {
            ID string `toml:"-"`
            Layout *Layout `toml:"-"`
            @{fields}
        }

        func (self *Layout) New@{e}(id string) *@{e} {
            if self.@{e} == nil {
                self.@{e} = make(map[string]*@{e})
            }
            if _, ok := self.@{e}[id]; ok {
                panic("@{e} already exists")
            }
            n := &@{e}{ID: id, Layout: self}
            self.@{e}[id] = n
            return n
        }

        @{functs}
        """

pts = t/''
for e, entity in sch.items():
    fields = t/''
    for f, field in entity.items():
        if field["type"] == "Ref":
            fields |= t/'v.REF@{f} = v.@{f}.ID'
        elif field["type"] == "Refs":
            fields |= t/"""
                v.REF@{f} = []string{}
                for _, p := range v.@{f} {
                    v.REF@{f} = append(v.REF@{f}, p.ID)
                }
                """
    pts |= t/"""
        for _, v := range self.@{e} {
            _ = v
            @{fields}
        }
        """

stp = t/''
for e, entity in sch.items():
    fields = t/''
    for f, field in entity.items():
        if field["type"] == "Ref":
            fields |= t/'v.@{f} = self.@{f}[v.REF@{f}]'
        elif field["type"] == "Refs":
            fields |= t/"""
            v.@{f} = []*@{field["refentity"]}{}
            for _, p := range v.REF@{f} {
                v.@{f} = append(v.@{f}, self.@{f}[p])
            }
            """
    stp |= t/"""
        for k, v := range self.@{e} {
            v.Layout = self
            v.ID = k
            @{fields}
        }
        """

out |= t%'' | t/"""
    func (self *Layout) pointersToStrings() {
        @{pts}
    }

    func (self *Layout) stringsToPointers() {
        @{stp}
    }
    """

with open("schema.go", 'w') as f:
    f.write(str(out))

###################################################################################################
# Python schema generator
###################################################################################################

classes = t/''
for e in sch:
    fields = t/''
    prints = t/''
    for fname, field in sch[e].items():
        if field["type"] == "Ref":
            fields |= t/'@{fname} = None'
            prints |= t/"""
                if self.@{fname} is not None:
                    iprint(level, "@{fname} = %s" % self.@{fname}.ID)
                """
        elif field["type"] == "Refs":
            fields |= t/'@{fname} = []'
            prints |= t/"""
                iprint(level, "@{fname} = [")
                for ent in self.@{fname}:
                    iprint(level + 1, ent.ID)
                iprint(level, "]")
                """
        else:
            fields |= t/'@{fname} = None'
            prints |= t/"""
                if self.@{fname} is not None:
                    iprint(level, "@{fname} = %s" % self.@{fname})
                """
    classes |= t%'' |  t/"""
        class @{e}Class(ProdspecEntity):
            ID = None
            @{fields}

            def print(self, level):
                iprint(level, "ID = %s" % self.ID)
                @{prints}
        """

out = t/''

for e in sch:
    out |= t%'' |  t/"""
        global @{e}
        @{e} = {}
        for id, obj in root["@{e}"].items():
            c = @{e}Class()
            c.ID = id
            for k, v in obj.items():
                setattr(c, k, v)
            @{e}[id] = c
        """

for ename, fields in sch.items():
    flds = t/''
    for fname, field in fields.items():
        if field["type"] == 'Ref':
            flds |= t/"""
               v.@{fname} = @{field["refentity"]}[v.@{fname}]
               """
        if field["type"] == 'Refs':
            flds |= t/"""
               for i in range(len(v.@{fname})):
                   v.@{fname}[i] = @{field["refentity"]}[v.@{fname}[i]]
               """
    out |= t%'' | t/"""
        for _, v in @{ename}.items():
            @{flds}
        """

out = t/"""
    #!/usr/bin/env python3

    from collections import namedtuple
    import sys
    import toml

    def iprint(level, s):
        print(("    " * level) + s)

    class ProdspecEntity:
        pass
    @{classes}

    def Load(filename):
        with open(filename, "r") as f:
            content = f.read()
        root =  toml.loads(content)
        @{out}

    Load("prodspec.toml")

    def pprint(obj, level):
       if isinstance(obj, ProdspecEntity):
           obj.print(level)
       elif isinstance(obj, dict):
           iprint(level, "{")
           for k, v in obj.items():
               iprint(level, k + ":")
               pprint(v, level + 1)
           iprint(level, "}")
       elif isinstance(obj, list):
           iprint(level, "[")
           for v in obj:
               pprint(v, level + 1)
               iprint(level, ",")
           iprint(level, "]")
       else:
           iprint(level, obj)

    pprint(eval(sys.argv[1]), 0)

    """

with open("../../../bin/query", 'w') as f:
    f.write(str(out))
