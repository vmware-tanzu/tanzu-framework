import { VmwCopyToClipboardButtonComponent } from 'src/app/shared/components/copy-to-clipboard-button/copy-to-clipboard-button.component';
import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { PreviewConfigComponent } from './preview-config.component';
import { SharedModule } from 'src/app/shared/shared.module';

describe('PreviewConfigComponent', () => {
  let component: PreviewConfigComponent;
  let fixture: ComponentFixture<PreviewConfigComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ PreviewConfigComponent, VmwCopyToClipboardButtonComponent ],
      imports: [SharedModule]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(PreviewConfigComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
