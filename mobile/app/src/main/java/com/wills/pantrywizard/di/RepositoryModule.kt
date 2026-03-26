package com.wills.pantrywizard.di

import com.wills.pantrywizard.data.remote.api.AiAgentApi
import com.wills.pantrywizard.data.remote.api.PantryApi
import com.wills.pantrywizard.data.repository.AiRepositoryImpl
import com.wills.pantrywizard.data.repository.PantryRepositoryImpl
import com.wills.pantrywizard.domain.repository.AiRepository
import com.wills.pantrywizard.domain.repository.PantryRepository
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import okhttp3.OkHttpClient
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import javax.inject.Named
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object RepositoryModule {

    @Provides
    @Singleton
    fun provideAiAgentApi(@Named("AuthOkHttpClient") okHttpClient: OkHttpClient): AiAgentApi {
        return Retrofit.Builder()
            .baseUrl("http://localhost:8000/")
            .client(okHttpClient)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
            .create(AiAgentApi::class.java)
    }

    @Provides
    @Singleton
    fun provideAiRepository(api: AiAgentApi): AiRepository {
        return AiRepositoryImpl(api)
    }

    @Provides
    @Singleton
    fun providePantryApi(@Named("AuthOkHttpClient") okHttpClient: OkHttpClient): PantryApi {
        return Retrofit.Builder()
            .baseUrl("http://your-go-backend:8080/") // Your Go Service URL
            .client(okHttpClient)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
            .create(PantryApi::class.java)
    }

    @Provides
    @Singleton
    fun providePantryRepository(api: PantryApi): PantryRepository {
        return PantryRepositoryImpl(api)
    }
}