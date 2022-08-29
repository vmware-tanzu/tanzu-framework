import {Injectable} from "@angular/core";
import {Observable} from "rxjs";
import {take} from "rxjs/operators";
import FileSaver from 'file-saver';

@Injectable({
    providedIn: 'root'
})
export class ExportService {
    export(backendResponse: Observable<string>,
           onFailure: (nameFile: string) => void) {
        backendResponse.pipe(take(1)).subscribe(
            ((data) => {
                const blob = new Blob([data], {type: "text/plain;charset=utf-8"});
                FileSaver.saveAs(blob, 'config.yaml');
            }),
            ((err) => {
                onFailure('Error encountered while creating export file: ' + err.toString());
            })
        )
    }
}
