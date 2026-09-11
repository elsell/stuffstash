import {expect,it} from 'vitest';
import {assetId,type AssetSummary} from '../../domain/assets/AssetSummary';
import {toAssetCardViewModel} from './AssetViewModels';
it('preserves expiration precision in cards used for search and conversation previews',()=>{
 const asset:AssetSummary={id:assetId('bottle'),title:'Bottle',kind:'item',lifecycleState:'active',locationLabel:'Closet',locationTrail:[],parentLocationTrail:[],description:'',updatedAtLabel:'Today',hasPhoto:false,expiration:{date:'2028-02',precision:'month'}};
 expect(toAssetCardViewModel(asset).expiration).toEqual(asset.expiration);
 expect(toAssetCardViewModel({...asset,expiration:undefined}).expiration).toBeUndefined();
});
