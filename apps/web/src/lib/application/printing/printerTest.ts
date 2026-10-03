import {PrintingFailure,type LabelSelection,type PrintJob,type PrintScope} from '$lib/domain/printing';
import type {PrinterTestIntent,PrintingRepository} from '$lib/ports/printingRepository';

export class PrinterTestRequest implements PrinterTestIntent {
    locked=false;
    selection:LabelSelection|null=null;
    result:PrintJob|null=null;
    private pending:Promise<PrintJob>|null=null;
    private ambiguous=false;
    constructor(private readonly repository:PrintingRepository,private readonly scope:PrintScope,private readonly printerId:string,private readonly key:string){}
    submit(selection:LabelSelection):Promise<PrintJob>{
        if(selection.printerId!==this.printerId||selection.copies!==1)
            return Promise.reject(new PrintingFailure('invalid'));
        if(this.locked&&JSON.stringify(selection)!==JSON.stringify(this.selection))
            return Promise.reject(new PrintingFailure('conflict'));
        if(this.result)return Promise.resolve(this.result);
        if(this.pending)return this.pending;
        this.selection={...selection};this.locked=true;
        this.pending=this.repository.testPrinter(this.scope,this.printerId,this.selection,this.key).then(result=>{this.result=result;return result;}).catch(error=>{
            const definite=error instanceof PrintingFailure&&['invalid','conflict','denied','authentication_required'].includes(error.kind);
            if(!definite)this.ambiguous=true;
            if(!this.ambiguous)this.locked=false;
            throw error;
        }).finally(()=>{this.pending=null;});
        return this.pending;
    }
}
