import { Component, OnInit } from '@angular/core'
import { MarkdownService } from 'ngx-markdown'

@Component({
  selector: 'ana-config',
  template: `<markdown [src]="configPath"></markdown>`,
})
export class ConfigComponent implements OnInit {

  configPath = 'doc/sig.md'

  constructor(private markdownService: MarkdownService) { }

  ngOnInit() {
    this.markdownService.renderer.link = (href: string, title: string, text: string) => {
      return '<a href="config' + href + '">' + text + '</a>'
    }
  }
}
