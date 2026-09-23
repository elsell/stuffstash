export const size = (width: number, height: number) => ({ type: 'size', width, height });
export const selectable = (selected: boolean, onClick: () => void, role: string) => ({ type: 'selectable', selected, onClick, role });
