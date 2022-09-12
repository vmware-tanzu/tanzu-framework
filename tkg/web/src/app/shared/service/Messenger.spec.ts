import { TestBed, async } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { HttpClientModule } from '@angular/common/http';

import { Messenger } from './Messenger';

describe('Messenger', () => {
    let service: Messenger;

    beforeEach(() => TestBed.configureTestingModule({
        imports: [
            HttpClientTestingModule
        ],
        providers: [
            Messenger
        ]
    }));

    beforeEach(() => {
        service = TestBed.get(Messenger);
    });

    it('should be created', () => {
        expect(service).toBeTruthy();
    });
});
