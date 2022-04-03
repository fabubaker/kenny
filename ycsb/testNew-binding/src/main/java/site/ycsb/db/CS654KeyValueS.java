package site.ycsb.db;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import site.ycsb.*;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.*;
import java.util.stream.Collectors;

public class CS654KeyValueS extends DB {

    HttpClient client;
    ObjectMapper mapper;

    public CS654KeyValueS(){
    }

    @Override
    public void init() throws DBException {
        // TODO Auto-generated method stub
        client = HttpClient.newHttpClient();
        mapper = new ObjectMapper();

    }

    @Override
    public Status read(String table, String key, Set<String> fields, Map<String, ByteIterator> result) {

        HttpResponse<String> response= getKV(key,fields,result);
        if(response!= null){
            try {
                Map<String, String> m = mapper.readValue(response.body(), Map.class);
                fillByteMap(m,result);
            } catch (JsonProcessingException e) {
                e.printStackTrace();
            }
            System.out.println(response.statusCode());
            return new Status("read","successful");
        }
        else{
            return new Status("read","failed: may be nothing to read");
        }
    }

    public HttpResponse<String> getKV(String key, Set<String> fields, Map<String, ByteIterator> result){

        if(fields == null || fields.isEmpty()){return null;}

        ArrayList<String> fieldList = new ArrayList<>();
        fieldList.addAll(fields);

        String json = null;
        try {
            json = mapper.writeValueAsString(fieldList);
        } catch (JsonProcessingException e) {
            e.printStackTrace();
        }

        HttpRequest request = HttpRequest.newBuilder()
                .POST(HttpRequest.BodyPublishers.ofString(json))
                .uri(URI.create(String.format("http://localhost:8080/%s", key)))
                .build();

        HttpResponse<String> response = null;
        try {
            response =  client.send(request, HttpResponse.BodyHandlers.ofString());
        } catch (IOException e) {
            e.printStackTrace();
        } catch (InterruptedException e) {
            e.printStackTrace();
        }
        return response;
    }


    @Override
    public Status scan(String table, String startkey, int recordcount, Set<String> fields, Vector<HashMap<String, ByteIterator>> result) {
        return null;
    }

    @Override
    public Status update(String table, String key, Map<String, ByteIterator> values) {
        return insert(table,key,values);
    }

    @Override
    public Status insert(String table, String key, Map<String, ByteIterator> values) {
        HttpResponse<String> response= putKV(key,values);

        if(response!= null){
            System.out.println(response.statusCode());
            return new Status("insert","successful");
        }
        else{
            return new Status("insert","failed:");
        }

    }

    public static Map<String, String> copyToStringValueMap (Map<String, ByteIterator> input) {
        Map<String, String> ret = new HashMap<>();
        for (Map.Entry<String, ByteIterator> entry : input.entrySet()) {
            ret.put(entry.getKey(), entry.getValue().toString());
        }
        return ret;
    }

    public static int fillByteMap (Map<String, String> input, Map<String, ByteIterator> result) {
        for (Map.Entry<String, String> entry : input.entrySet()) {
            result.put(entry.getKey(), new ByteArrayByteIterator(entry.getValue().getBytes()));
        }
        return 0;
    }

    public HttpResponse<String> putKV(String key, Map<String, ByteIterator> values){

        String json = null;

        try {
            json = mapper.writeValueAsString(copyToStringValueMap(values));
        } catch (JsonProcessingException e) {
            e.printStackTrace();
        }
        HttpRequest request = HttpRequest.newBuilder()
                .PUT(HttpRequest.BodyPublishers.ofString(json))
                .uri(URI.create(String.format("http://localhost:8080/%s", key)))
                .build();

        HttpResponse<String> response = null;
        try {
            response =  client.send(request, HttpResponse.BodyHandlers.ofString());
        } catch (IOException e) {
            e.printStackTrace();
        } catch (InterruptedException e) {
            e.printStackTrace();
        }

        return response;
    }

    @Override
    public Status delete(String table, String key) {
        return null;
    }


}
