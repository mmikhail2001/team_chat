import { UserOBJ } from "./models";

// TODO: что за тип? почему отношение наследуется от user? 
// видимо все друзья
// type: друг, не друг
export interface Relationship extends UserOBJ {
    type: number
}
