const modifier=(type:string)=>(value:unknown)=>({type,value});
export const accessibilityLabel=modifier('accessibilityLabel');
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
