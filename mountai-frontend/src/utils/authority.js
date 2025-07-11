import { reloadAuthorized } from './Authorized'; // use localStorage to store the authority info, which might be sent from server in actual project.

export function getAuthority(str) {
  const authorityString =
    typeof str === 'undefined' && localStorage ? localStorage.getItem('ams-user-authority') : str; // authorityString could be admin, "admin", ["admin"]

  let authority;



  try {
    if (authorityString) {
      authority = JSON.parse(authorityString);
    }
  } catch (e) {
    authority = authorityString;
  }

  if (typeof authority === 'string') {
    return [authority];
  } // preview.pro.ant.design only do not use in your production.
  // preview.pro.ant.design 专用环境变量，请不要在你的项目中使用它。

  if (!authority && ANT_DESIGN_PRO_ONLY_DO_NOT_USE_IN_YOUR_PRODUCTION === 'site') {
    return ['admin'];
  }
 
  if(!authority) {
    return ['guest'];
  } else {
    return ['admin'];
  } 
  // 为了传['admin']供鉴权
  //  return authority.roles;
}
export function setAuthority(authority) {
//  console.log('设置token==>\n', authority);
 // payload
  const proAuthority = typeof authority === 'string' ? [authority] : authority;
  // console.log('authority', authority, proAuthority);
  if(authority && authority !== undefined) {
    localStorage.clear();
    localStorage.setItem('ams-user-authority', JSON.stringify(proAuthority)); // auto reload
  }


  reloadAuthorized();
}
