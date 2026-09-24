const modifier=(type:string)=>(value:unknown)=>({type,value});
export const accessibilityLabel=modifier('accessibilityLabel');
export const accessibilityHint=modifier('accessibilityHint');
export const buttonStyle=modifier('buttonStyle');
export const controlSize=modifier('controlSize');
export const disabled=modifier('disabled');
export const fixedSize=modifier('fixedSize');
export const frame=modifier('frame');
export const tint=modifier('tint');
export const pickerStyle=modifier('pickerStyle');
export const tag=modifier('tag');

export const labelStyle=modifier('labelStyle');

export const labelsHidden=modifier('labelsHidden');
export const textFieldStyle=modifier('textFieldStyle');

export const keyboardType=modifier('keyboardType');
export const textContentType=modifier('textContentType');
export const textInputAutocapitalization=modifier('textInputAutocapitalization');
export const autocorrectionDisabled=(value=true)=>({type:'autocorrectionDisabled',value});
export const accessibilityValue=modifier('accessibilityValue');
export const contentShape=modifier('contentShape');
export const shapes={rectangle:()=>({type:'rectangle'})};

export const font=modifier('font');
export const foregroundStyle=modifier('foregroundStyle');
export const listStyle=modifier('listStyle');
