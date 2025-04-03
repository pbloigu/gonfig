/* tslint:disable */
/* eslint-disable */

type readonlyP = { readonly?: '__readonly' };
type writeonlyP = { writeonly?: '__writeonly' };
type Primitive = string | Function | number | boolean | Symbol | undefined | null | Date;
type PropsWithoutReadonly<T> = {
    [key in keyof T]: T[key] extends readonlyP
    ? NonNullable<T[key]['readonly']> extends '__readonly' ? never : key
    : key
}[keyof T];
export type WithoutReadonly<T> = T extends any ?
    T extends Primitive ? T :
    T extends Array<infer U> ? WithoutReadonly<U>[] :
    keyof T extends never ? unknown :
    {
        [key in keyof Pick<T, PropsWithoutReadonly<T>>]: Pick<T, PropsWithoutReadonly<T>>[key] extends any
        ? WithoutReadonly<Pick<T, PropsWithoutReadonly<T>>[key]>
        : never
    } : never;
type PropsWithoutWriteonly<T> = {
    [key in keyof T]: T[key] extends writeonlyP
    ? NonNullable<T[key]['writeonly']> extends '__writeonly' ? never : key
    : key
}[keyof T];
export type WithoutWriteonly<T> = T extends any ?
    T extends Primitive ? T :
    T extends Array<infer U> ? WithoutWriteonly<U>[] :
    keyof T extends never ? unknown :
    {
        [key in keyof Pick<T, PropsWithoutWriteonly<T>>]: Pick<T, PropsWithoutWriteonly<T>>[key] extends any
        ? WithoutWriteonly<Pick<T, PropsWithoutWriteonly<T>>[key]>
        : never
    } : never;
export type Application = {
    readonly $schema?: (string) & readonlyP;
    readonly apiKey?: (string) & readonlyP;
    configuration?: Configuration;
    readonly hostname?: (string) & readonlyP;
    readonly id?: (string) & readonlyP;
    readonly ip?: (string) & readonlyP;
    measurements?: string[];
    name: string;
};
export type Configuration = {
    readonly $schema?: (string) & readonlyP;
    data: string;
    readonly time?: (string) & readonlyP;
};
export type ErrorDetail = {
    location?: string;
    message?: string;
    value?: unknown;
};
export type ErrorModel = {
    readonly $schema?: (string) & readonlyP;
    detail?: string;
    errors?: ErrorDetail[];
    instance?: string;
    status?: number;
    title?: string;
    type?: string;
};
export type LoginRequest = {
    readonly $schema?: (string) & readonlyP;
    password: string;
    username: string;
};
export type LoginResponse = {
    readonly $schema?: (string) & readonlyP;
    expiry: number;
    token: string;
};
export type Measurement = {
    lastValue?: MeasurementValue;
    name: string;
};
export type MeasurementValue = {
    data: string;
    readonly time?: (string) & readonlyP;
};
export type MeasurementValues = {
    readonly $schema?: (string) & readonlyP;
    measurement: Measurement;
    page: number;
    pageSize: number;
    total: number;
    values: MeasurementValue[];
};
