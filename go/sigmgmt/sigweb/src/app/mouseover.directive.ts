import { Directive, EventEmitter, HostListener, Output } from '@angular/core'

@Directive({
  selector: '[anaMouseover]'
})
export class MouseoverDirective {
  @Output() mouseover = new EventEmitter<boolean>()

  @HostListener('mouseenter') onMouseEnter() {
    this.mouseover.emit(true)
  }

  @HostListener('mouseleave') onMouseLeave() {
    setTimeout(() => {
      this.mouseover.emit(false)
    }, 200)
  }
}
